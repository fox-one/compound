package rest

import (
	"compound/handler/render"
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Handle handle rest request
func Handle(ctx context.Context) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	router.Get("/markets", nil)
	router.Get("/markets/{symbol}", nil)

	router.Post("/supply", nil)
	router.Post("/borrow", nil)
	router.Post("/redeemsupply", nil)
	router.Post("/repayborrow", nil)

	router.Get("/liquidities", nil)

	return router
}
