package run

import (
	"net/http"

	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/{metricName}/{value}", handlers.SetGaugeMetric)
			r.Post("/{}", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/{metricName}/{value}", handlers.SetCounterMetric)
			r.Post("/{}", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})
		})
		// fixme maybe rewrite r.NotFound to 501 code?
		r.Route("/{otherType}/", func(r chi.Router) {
			r.Post("/{metricName}", handlers.NotImplemented)
			r.Post("/{metricName}/{value}", handlers.NotImplemented)
		})
		r.Post("/", handlers.SetMetricJSON)
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", handlers.GetMetric)
		r.Post("/", handlers.GetMetricJSON)
	})
	r.Get("/", handlers.ListMetrics)
	return r
}
