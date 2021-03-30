package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle price proposal
func (w *Payee) handleProposalProvidePriceEvent(ctx context.Context, output *core.Output, member *core.Member, traceID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "handleProposalProvidePriceEvent")
	var data proposal.ProvidePriceReq
	if _, e := mtg.Scan(body, &data); e != nil {
		return nil
	}

	blockNum := core.CalculatePriceBlock(output.CreatedAt)

	market, isRecordNotFound, e := w.marketStore.FindBySymbol(ctx, data.Symbol)
	if isRecordNotFound {
		return nil
	}

	if e != nil {
		return e
	}

	log.Infof("asset:%s,block:%d, output.updated_at:%v", data.Symbol, blockNum, output.CreatedAt)

	price, isRecordNotFound, e := w.priceStore.FindByAssetBlock(ctx, market.AssetID, blockNum)
	if e != nil {
		if isRecordNotFound {
			// new one
			ts := make(map[string]decimal.Decimal)
			ts[member.ClientID] = data.Price

			bs, e := json.Marshal(ts)
			if e != nil {
				return e
			}

			price = &core.Price{
				AssetID:     market.AssetID,
				BlockNumber: blockNum,
				Content:     bs,
			}
			if e = w.priceStore.Create(ctx, price); e != nil {
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
	tickerMap := make(map[string]decimal.Decimal)
	if e = json.Unmarshal(price.Content, &tickerMap); e != nil {
		return e
	}

	_, found := tickerMap[member.ClientID]
	if !found {
		tickerMap[member.ClientID] = data.Price
	}

	priceLen := len(tickerMap)
	if priceLen < int(w.system.Threshold) {
		// less than threshold, update content
		// marshal tickers content
		bs, e := json.Marshal(tickerMap)
		if e != nil {
			return e
		}

		price.Content = bs
		if e = w.priceStore.Update(ctx, price, output.ID); e != nil {
			log.WithError(e).Errorln("update price err")
			return e
		}

		return nil
	}

	sumOfPrice := decimal.Zero
	for _, t := range tickerMap {
		sumOfPrice = sumOfPrice.Add(t)
	}

	avgPrice := sumOfPrice.Div(decimal.NewFromInt(int64(priceLen)))

	validPrices := make(map[string]decimal.Decimal)
	for k, v := range tickerMap {
		delta := v.Sub(avgPrice).Abs()
		changeRate := delta.Div(avgPrice)
		if changeRate.LessThan(decimal.NewFromFloat(0.05)) {
			validPrices[k] = v
		}
	}

	lenOfValidPrices := len(validPrices)
	if len(validPrices) < int(w.system.Threshold) {
		// less than threshold, update content
		// marshal tickers content
		bs, e := json.Marshal(tickerMap)
		if e != nil {
			return e
		}

		price.Content = bs
		if e = w.priceStore.Update(ctx, price, output.ID); e != nil {
			log.WithError(e).Errorln("update price err")
			return e
		}

		return nil
	}

	// avg price
	sumOfPrice = decimal.Zero
	for _, p := range validPrices {
		sumOfPrice = sumOfPrice.Add(p)
	}
	price.Price = sumOfPrice.Div(decimal.NewFromInt(int64(lenOfValidPrices)))

	price.PassedAt = sql.NullTime{
		Time:  output.CreatedAt,
		Valid: true,
	}

	// marshal tickers content
	bs, e := json.Marshal(tickerMap)
	if e != nil {
		return e
	}

	price.Content = bs
	if e = w.priceStore.Update(ctx, price, output.ID); e != nil {
		log.WithError(e).Errorln("update price err")
		return e
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	market.Price = price.Price
	market.PriceUpdatedAt = output.CreatedAt
	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.WithError(e).Errorln("update market price err")
		return e
	}

	return nil
}
