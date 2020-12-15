package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleProposalProvidePriceEvent(ctx context.Context, output *core.Output, member *core.Member, traceID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "handleProposalProvidePriceEvent")
	var data proposal.ProvidePriceReq
	if _, e := mtg.Scan(body, &data); e != nil {
		return nil
	}

	blockNum, e := w.blockService.GetBlock(ctx, output.UpdatedAt)
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		priceTickers := make([]*core.PriceTicker, 0)
		price, isRecordNotFound, e := w.priceStore.FindByAssetBlock(ctx, data.AssetID, blockNum)
		if e != nil {
			if isRecordNotFound {
				// new one
				priceTickers = append(priceTickers, &core.PriceTicker{
					Provider: member.ClientID,
					Price:    data.Price,
				})

				bs, e := json.Marshal(priceTickers)
				if e != nil {
					return e
				}

				price = &core.Price{
					AssetID:     data.AssetID,
					BlockNumber: blockNum,
					Content:     bs,
				}
				if e = w.priceStore.Create(ctx, tx, price); e != nil {
					log.WithError(e).Errorln("create price err")
					return e
				}
				return nil
			}
			return e
		}

		if price.PassedAt.Valid {
			//passed
			return nil
		}

		// exists
		price.AssetID = data.AssetID
		price.BlockNumber = blockNum

		if e = json.Unmarshal(price.Content, &priceTickers); e != nil {
			return e
		}

		priceTickers = append(priceTickers, &core.PriceTicker{
			Provider: member.ClientID,
			Price:    data.Price,
		})

		priceLen := len(priceTickers)
		if priceLen < int(w.system.Threshold) {
			// less than threshold
			return nil
		}

		sort.Slice(priceTickers, func(i, j int) bool {
			return priceTickers[i].Price.LessThan(priceTickers[j].Price)
		})

		validPrices := make(map[string]decimal.Decimal)

		for i := 1; i < priceLen; i++ {
			first := priceTickers[i-1]
			second := priceTickers[i]
			delta := second.Price.Sub(first.Price).Abs()
			changeRatio := delta.Div(first.Price)
			if changeRatio.LessThan(decimal.NewFromFloat(0.05)) {
				validPrices[first.Provider] = first.Price
				validPrices[second.Provider] = first.Price
			}
		}

		lenOfValidPrices := len(validPrices)
		if len(validPrices) < int(w.system.Threshold) {
			// less than threshold
			return nil
		}

		sumOfPrice := decimal.Zero
		for _, p := range validPrices {
			sumOfPrice = sumOfPrice.Add(p)
		}
		price.Price = sumOfPrice.Div(decimal.NewFromInt(int64(lenOfValidPrices)))

		price.PassedAt = sql.NullTime{
			Time:  output.UpdatedAt,
			Valid: true,
		}
		bs, e := json.Marshal(priceTickers)
		if e != nil {
			return e
		}

		price.Content = bs
		if e = w.priceStore.Update(ctx, tx, price); e != nil {
			log.WithError(e).Errorln("update price err")
			return e
		}

		// update market
		market, isRecordNotFound, e := w.marketStore.Find(ctx, data.AssetID)
		if isRecordNotFound {
			return nil
		}

		if e != nil {
			return e
		}

		// accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.UpdatedAt); e != nil {
			return e
		}

		market.Price = price.Price
		market.PriceUpdatedAt = output.UpdatedAt
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.WithError(e).Errorln("update market price err")
			return e
		}

		return nil
	})
}
