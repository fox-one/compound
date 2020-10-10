package auth

import (
	"compound/core"
	"compound/handler/request"
	"net/http"
	"strings"

	"github.com/fox-one/pkg/logger"
)

// HandleAuthentication handle authentication
func HandleAuthentication(session core.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)

			accessToken := getBearerToken(r)
			if accessToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := session.Login(ctx, accessToken)
			if err != nil {
				next.ServeHTTP(w, r)
				log.WithError(err).Debugln("parse access token error:", err)
				return
			}

			next.ServeHTTP(w, r.WithContext(request.NewContext(ctx).WithUser(user)))
		}

		return http.HandlerFunc(fn)
	}
}

func getBearerToken(r *http.Request) string {
	s := r.Header.Get("Authorization")
	return strings.TrimPrefix(s, "Bearer ")
}
