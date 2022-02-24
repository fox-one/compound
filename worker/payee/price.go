package payee

import (
	"compound/core"
	"context"
	"encoding/base64"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/blst"
)

func (w *Payee) handlePriceEvent(ctx context.Context, output *core.Output, priceData *core.PriceData) error {
	log := logger.FromContext(ctx).WithField("worker", "handle_dirt_price_event")

	market, e := w.marketStore.Find(ctx, priceData.AssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return nil
	}

	// accrue interest
	if e = AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	if output.ID > market.Version {
		log.Infoln("dirt_price: asset:", priceData.AssetID, ", price:", priceData.Price, ",time:", output.CreatedAt)
		market.Price = priceData.Price
		market.PriceUpdatedAt = output.CreatedAt
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.WithError(e).Errorln("update market price err")
			return e
		}
	}

	return nil
}

func (w *Payee) decodePriceTransaction(ctx context.Context, businessData []byte) (*core.PriceData, error) {
	var p core.PriceData
	if err := p.UnmarshalBinary(businessData); err != nil {
		return nil, nil
	}

	ss, err := w.oracleSignerStore.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	market, err := w.marketStore.Find(ctx, p.AssetID)
	if err != nil {
		return nil, err
	} else if market.ID == 0 {
		return nil, nil
	}

	signers := make([]*core.Signer, len(ss))
	for idx, s := range ss {
		bts, err := base64.StdEncoding.DecodeString(s.PublicKey)
		if err != nil {
			return nil, nil
		}

		pub := blst.PublicKey{}
		if err := pub.FromBytes(bts); err != nil {
			return nil, nil
		}

		signers[idx] = &core.Signer{
			Index:     uint64(idx) + 1,
			VerifyKey: &pub,
		}
	}

	if verifyPriceData(&p, signers, market.PriceThreshold) {
		return &p, nil
	}

	return nil, nil
}

func verifyPriceData(p *core.PriceData, signers []*core.Signer, threshold int) bool {
	var pubs []*blst.PublicKey
	for _, signer := range signers {
		if p.Signature.Mask&(0x1<<signer.Index) != 0 {
			pubs = append(pubs, signer.VerifyKey)
		}
	}

	return len(pubs) >= threshold &&
		blst.AggregatePublicKeys(pubs).Verify(p.Payload(), &p.Signature.Signature)
}
