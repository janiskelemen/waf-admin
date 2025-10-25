package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "reqStart", time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if t, ok := r.Context().Value("reqStart").(time.Time); ok {
			log.Info().Str("method", r.Method).Str("path", r.URL.Path).Dur("dur", time.Since(t)).Msg("req")
		}
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(){ if rec := recover(); rec != nil { http.Error(w, "internal error", 500) } }()
		next.ServeHTTP(w, r)
	})
}
