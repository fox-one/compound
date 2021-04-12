package rest

import (
	"compound/core"
	"compound/handler/render"
	"fmt"
	"net/http"
	"time"

	"github.com/fox-one/pkg/uuid"
)

func priceRequestsHandler(system *core.System, marketStr core.IMarketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		markets, e := marketStr.All(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		var members []string
		for _, m := range system.Members {
			members = append(members, m.ClientID)
		}

		var requests []*core.PriceRequest
		for _, m := range markets {
			if time.Now().After(m.PriceUpdatedAt.Add(10 * time.Minute)) {
				requests = append(requests, &core.PriceRequest{
					TraceID: uuid.Modify(m.AssetID, fmt.Sprintf("price-request:%s:%d", system.ClientID, time.Now().Unix()/600)),
					Asset:   core.Asset{AssetID: m.AssetID, Symbol: m.Symbol},
					Receiver: &core.Receiver{
						Threshold: system.Threshold,
						Members:   members,
					},
					Signers:   system.PriceOracleSigners,
					Threshold: system.Threshold,
				})
			}
		}

		var response struct {
			Data interface{} `json:"data"`
		}

		response.Data = requests

		render.JSON(w, response)
	}
}
