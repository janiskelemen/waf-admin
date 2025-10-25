package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/janiskelemen/waf-admin/internal/auth"
)

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(requestID, logger, recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/health", s.health)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/api/openapi.yaml")
	})

	p := chi.NewRouter()
	p.Use(auth.Bearer(s.cfg.Auth.Token))

	p.Get("/v1/sites", s.listSites)
	p.Get("/v1/sites/{name}", s.getSite)
	p.Put("/v1/sites/{name}", s.putSite)
	p.Delete("/v1/sites/{name}", s.deleteSite)

	p.Get("/v1/rules/{site}", s.listRules)
	p.Get("/v1/rules/{site}/{file}", s.getRule)
	p.Put("/v1/rules/{site}/{file}", s.putRule)
	p.Delete("/v1/rules/{site}/{file}", s.deleteRule)

	p.Post("/v1/validate", s.validate)
	p.Post("/v1/apply", s.apply)
	p.Post("/v1/backup", s.backupNow)

	r.Mount("/", p)
	return r
}
