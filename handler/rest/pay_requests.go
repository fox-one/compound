package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"encoding/base64"
	"net/http"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

func payRequestsHandler(system *core.System, dapp *core.Wallet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			MemoBase64 string          `json:"memo_base64,omitempty"`
			AssetID    string          `json:"asset_id,omitempty"`
			Amount     decimal.Decimal `json:"amount,omitempty"`
			TraceID    string          `json:"trace_id,omitempty"`
			FollowID   string          `json:"follow_id,omitempty"`
		}

		if err := param.Binding(r, &params); err != nil {
			render.BadRequest(w, err)
			return
		}

		var followID []byte
		if follow, err := uuid.FromString(params.FollowID); err == nil && follow != uuid.Nil {
			followID = follow.Bytes()
		}

		data, err := base64.StdEncoding.DecodeString(params.MemoBase64)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		memoBytes, err := core.TransactionAction{FollowID: followID, Body: data}.Encode()
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		assetID := system.VoteAsset
		amount := system.VoteAmount
		if params.AssetID != "" {
			assetID = params.AssetID
		}
		if params.Amount.IsPositive() {
			amount = params.Amount
		}

		input := mixin.TransferInput{
			AssetID: assetID,
			Amount:  amount,
			TraceID: params.TraceID,
			Memo:    base64.StdEncoding.EncodeToString(memoBytes),
		}
		input.OpponentMultisig.Receivers = system.MemberIDs
		input.OpponentMultisig.Threshold = system.Threshold

		payment, err := dapp.Client.VerifyPayment(ctx, input)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		url := mixin.URL.Codes(payment.CodeID)

		var response struct {
			URL           string               `json:"url"`
			TransferInput *mixin.TransferInput `json:"transfer_input"`
		}

		response.URL = url
		response.TransferInput = &input

		render.JSON(w, response)
	}
}
