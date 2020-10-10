package auth

import (
	"net/http"

	"compound/core"
	"compound/handler/param"
	"compound/handler/render"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/twitchtv/twirp"
)

// HandleOauth handle oauth
func HandleOauth(mixinConfig *core.Mixin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Code string `json:"code,omitempty" valid:"minstringlength(6),required"`
		}

		if err := param.Binding(r, &body); err != nil {
			render.Error(w, err)
			return
		}

		ctx := r.Context()

		token, scope, err := mixin.AuthorizeToken(ctx, mixinConfig.ClientID, mixinConfig.ClientSecret, body.Code, "")
		if err != nil {
			render.Error(w, twirp.InvalidArgumentError("code", err.Error()))
			return
		}

		render.JSON(w, render.H{"token": token, "scope": scope})
	}
}
