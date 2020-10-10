package hc

import (
	"compound/handler/render"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Handle handle hc request
func Handle(ver string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	r.Handle("/", handle(ver))
	return r
}

func handle(version string) http.HandlerFunc {
	b := time.Now()
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(b).Truncate(time.Millisecond)
		render.JSON(w, render.H{
			"uptime":  uptime.String(),
			"version": version,
		})
	}
}
