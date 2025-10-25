package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/janiskelemen/waf-admin/internal/domain"
	"github.com/janiskelemen/waf-admin/internal/reload"
	"github.com/janiskelemen/waf-admin/internal/render"
	"github.com/janiskelemen/waf-admin/internal/storage"
)

var (
	siteNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	fileNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.conf(?:\.disabled)?$`)
)

type Server struct {
	cfg    *Config
	store  storage.Storage
	driver render.Driver
	rel    reload.Reloader
	http   *http.Server
}

func NewServer(cfg *Config, st storage.Storage, dr render.Driver, rl reload.Reloader) *Server {
	return &Server{cfg: cfg, store: st, driver: dr, rel: rl}
}

func (s *Server) Start() error {
	s.http = &http.Server{Addr: s.cfg.Server.Bind, Handler: s.routes()}
	log.Info().Str("bind", s.cfg.Server.Bind).Msg("starting waf-admin")
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error { return s.http.Shutdown(ctx) }

func (s *Server) health(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") }

func writeJSON(w http.ResponseWriter, v any, err error) {
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) { http.Error(w, msg, code) }

func (s *Server) listSites(w http.ResponseWriter, r *http.Request) {
	sites, err := domain.ListSites(r.Context(), s.driver, s.store)
	writeJSON(w, sites, err)
}

func (s *Server) getSite(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "name")
	if !siteNameRe.MatchString(site) {
		writeErr(w, 400, "invalid site name")
		return
	}
	path := filepath.Join(s.driver.LayoutSites(), site+".caddy")
	b, err := s.store.Read(r.Context(), path)
	if err != nil {
		writeErr(w, 404, "site not found")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write(b)
}

type putSiteReq struct {
	Content string `json:"content"`
}

func (s *Server) putSite(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "name")
	if !siteNameRe.MatchString(site) {
		writeErr(w, 400, "invalid site name")
		return
	}
	var req putSiteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeErr(w, 400, "invalid content")
		return
	}
	path := filepath.Join(s.driver.LayoutSites(), site+".caddy")
	if err := s.store.WriteAtomic(r.Context(), path, []byte(req.Content), 0o644); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if err := s.store.MkdirAll(r.Context(), filepath.Join(s.driver.LayoutRulesRoot(), site, "rules"), 0o755); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if err := s.applyNow(r.Context()); err != nil {
		writeErr(w, 400, "validate/apply failed: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) deleteSite(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "name")
	if !siteNameRe.MatchString(site) {
		writeErr(w, 400, "invalid site name")
		return
	}
	path := filepath.Join(s.driver.LayoutSites(), site+".caddy")
	_ = s.store.Delete(r.Context(), path)
	if err := s.applyNow(r.Context()); err != nil {
		writeErr(w, 400, "validate/apply failed: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "site")
	if !siteNameRe.MatchString(site) {
		writeErr(w, 400, "invalid site name")
		return
	}
	dir := filepath.Join(s.driver.LayoutRulesRoot(), site, "rules")
	ents, err := s.store.List(r.Context(), dir)
	if err != nil {
		writeErr(w, 404, "no rules")
		return
	}
	files := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if fileNameRe.MatchString(name) {
			files = append(files, name)
		}
	}
	writeJSON(w, files, nil)
}

func (s *Server) getRule(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "site")
	file := chi.URLParam(r, "file")
	if !siteNameRe.MatchString(site) || !fileNameRe.MatchString(file) {
		writeErr(w, 400, "invalid name")
		return
	}
	path := filepath.Join(s.driver.LayoutRulesRoot(), site, "rules", file)
	b, err := s.store.Read(r.Context(), path)
	if err != nil {
		writeErr(w, 404, "rule not found")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write(b)
}

type putRuleReq struct {
	Content string `json:"content"`
}

func (s *Server) putRule(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "site")
	file := chi.URLParam(r, "file")
	if !siteNameRe.MatchString(site) || !fileNameRe.MatchString(file) {
		writeErr(w, 400, "invalid name")
		return
	}
	var req putRuleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeErr(w, 400, "invalid content")
		return
	}
	path := filepath.Join(s.driver.LayoutRulesRoot(), site, "rules", file)
	if err := s.store.MkdirAll(r.Context(), filepath.Dir(path), 0o755); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if err := s.store.WriteAtomic(r.Context(), path, []byte(req.Content), 0o644); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if err := s.applyNow(r.Context()); err != nil {
		writeErr(w, 400, "validate/apply failed: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) deleteRule(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "site")
	file := chi.URLParam(r, "file")
	if !siteNameRe.MatchString(site) || !fileNameRe.MatchString(file) {
		writeErr(w, 400, "invalid name")
		return
	}
	path := filepath.Join(s.driver.LayoutRulesRoot(), site, "rules", file)
	_ = s.store.Delete(r.Context(), path)
	if err := s.applyNow(r.Context()); err != nil {
		writeErr(w, 400, "validate/apply failed: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	if err := s.driver.Validate(r.Context()); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) apply(w http.ResponseWriter, r *http.Request) {
	if err := s.applyNow(r.Context()); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true}, nil)
}

func (s *Server) applyNow(ctx context.Context) error {
	if err := s.driver.Validate(ctx); err != nil {
		return err
	}
	if err := s.rel.Reload(ctx); err != nil {
		return err
	}
	return nil
}
