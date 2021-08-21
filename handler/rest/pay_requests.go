package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/pkg/mtg"
	"crypto/ed25519"
	"encoding/base64"
	"net/http"

	"github.com/fox-one/mixin-sdk-go"
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
			WithGas    bool            `json:"with_gas,omitempty"`
		}

		if err := param.Binding(r, &params); err != nil {
			render.BadRequest(w, err)
			return
		}

		memoBytes, err := base64.StdEncoding.DecodeString(params.MemoBase64)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		key := mixin.GenerateEd25519Key()
		pub := system.PrivateKey.Public().(ed25519.PublicKey)

		memoEncrypt, err := mtg.Encrypt(memoBytes, key, pub)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		assetID := params.AssetID
		amount := params.Amount

		if params.WithGas {
			assetID = system.VoteAsset
			amount = system.VoteAmount
		}

		input := mixin.TransferInput{
			AssetID: assetID,
			Amount:  amount,
			TraceID: params.TraceID,
			Memo:    base64.StdEncoding.EncodeToString(memoEncrypt),
		}
		input.OpponentMultisig.Receivers = system.MemberIDs()
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
