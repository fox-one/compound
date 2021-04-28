package rest

import (
	"compound/core"
	"compound/handler/render"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/fox-one/pkg/uuid"
	"github.com/pandodao/blst"
)

// TODO: stop price scan
func priceRequestsHandler(system *core.System, marketStr core.IMarketStore, oracleSignerStr core.OracleSignerStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		markets, e := marketStr.All(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		// members
		var members []string
		for _, m := range system.Members {
			members = append(members, m.ClientID)
		}

		// signers
		ss, err := oracleSignerStr.FindAll(ctx)
		if err != nil {
			render.BadRequest(w, e)
			return
		}

		signers := make([]*core.Signer, len(ss))
		for idx, s := range ss {
			bts, err := base64.StdEncoding.DecodeString(s.PublicKey)
			if err != nil {
				render.BadRequest(w, e)
				return
			}

			pub := blst.PublicKey{}
			if err := pub.FromBytes(bts); err != nil {
				render.BadRequest(w, e)
				return
			}

			signers[idx] = &core.Signer{
				Index:     uint64(idx) + 1,
				VerifyKey: &pub,
			}
		}

		requests := make([]*core.PriceRequest, 0)
		for _, m := range markets {
			if time.Now().After(m.PriceUpdatedAt.Add(10 * time.Minute)) {
				requests = append(requests, &core.PriceRequest{
					TraceID: uuid.Modify(m.AssetID, fmt.Sprintf("price-request:%s:%d", system.ClientID, time.Now().Unix()/600)),
					Asset:   core.Asset{AssetID: m.AssetID, Symbol: m.Symbol},
					Receiver: &core.Receiver{
						Threshold: system.Threshold,
						Members:   members,
					},
					Signers:   signers,
					Threshold: system.PriceThreshold,
				})
			}
		}

		var response struct {
			Data []*core.PriceRequest `json:"data"`
		}

		response.Data = requests

		render.JSON(w, response)
	}
}
