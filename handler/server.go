package handler

import (
	"compound/core"
	"compound/handler/render"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Server server
type Server struct {
	cfg *core.Config
}

// New new server function
func New(
	cfg *core.Config,
) Server {
	return Server{
		cfg: cfg,
	}
}

// HandleRestAPI handle restful apis
func (s Server) HandleRestAPI() http.Handler {
	r := chi.NewRouter()
	r.Use(resetRoutePath)
	r.Use(render.WrapResponse(true))
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	// rt := reversetwirp.NewSingleTwirpServerProxy()
	// r.Post("/oauth", auth.HandleOauth(&s.cfg.Mixin))

	return r
}

func resetRoutePath(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if c := chi.RouteContext(ctx); c != nil {
			c.RoutePath = r.URL.Path
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
