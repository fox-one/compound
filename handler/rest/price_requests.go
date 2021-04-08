package rest

import (
	"compound/core"
	"compound/handler/render"
	"fmt"
	"net/http"
	"time"

	"github.com/fox-one/pkg/uuid"
)

func priceRequestsHandler(cfg *core.Config, marketStr core.IMarketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		markets, e := marketStr.All(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		var members []string
		for _, m := range cfg.Group.Members {
			members = append(members, m.ClientID)
		}

		var requests []*core.PriceRequest
		for _, m := range markets {
			requests = append(requests, &core.PriceRequest{
				TraceID: uuid.Modify(m.AssetID, fmt.Sprintf("price-request:%s:%d", cfg.Dapp.ClientID, time.Now().Unix()/600)),
				Asset:   core.Asset{AssetID: m.AssetID, Symbol: m.Symbol},
				Receiver: &core.Receiver{
					Threshold: cfg.Group.Threshold,
					Members:   members,
				},
				Signers:   cfg.PriceOracle.Signers,
				Threshold: cfg.Group.Threshold,
			})
		}

		render.JSON(w, requests)
	}
}
