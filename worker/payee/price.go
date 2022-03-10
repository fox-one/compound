package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"encoding/base64"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/blst"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handlePriceEvent(ctx context.Context, output *core.Output) error {
	var priceData core.PriceData
	if err := compound.Require(
		priceData.UnmarshalBinary(w.decodeMemo(output.Memo)) == nil,
		"payee/invalid-price-data",
	); err != nil {
		return err
	}

	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"event":     "price-oracle",
		"asset":     priceData.AssetID,
		"price":     priceData.Price,
		"timestamp": time.Unix(priceData.Timestamp, 0),
	})
	ctx = logger.WithContext(ctx, log)

	market, err := w.mustGetMarket(ctx, priceData.AssetID)
	if err != nil {
		return err
	}

	if market.Version >= output.ID {
		return nil
	}

	ss, err := w.oracleSignerStore.FindAll(ctx)
	if err != nil {
		log.WithError(err).Errorln("oracles.FindAll")
		return err
	}

	signers := make([]*core.Signer, len(ss))
	for idx, s := range ss {
		bts, err := base64.StdEncoding.DecodeString(s.PublicKey)
		if e := compound.Require(
			err == nil,
			"payee/invalid-oracle-signer",
		); e != nil {
			return e
		}

		pub := blst.PublicKey{}
		if e := compound.Require(
			pub.FromBytes(bts) == nil,
			"payee/invalid-oracle-signer",
		); e != nil {
			return e
		}

		signers[idx] = &core.Signer{
			Index:     uint64(idx) + 1,
			VerifyKey: &pub,
		}
	}

	if err := compound.Require(
		verifyPriceData(&priceData, signers, market.PriceThreshold),
		"payee/oracle-verify-failed",
	); err != nil {
		log.WithError(err).Errorln("price data verify failed")
		return err
	}

	market.Price = priceData.Price
	market.PriceUpdatedAt = output.CreatedAt
	AccrueInterest(ctx, market, output.CreatedAt)
	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("update market price err")
		return err
	}

	log.Infoln("price updated")
	return nil
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
