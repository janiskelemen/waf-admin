package scheduler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/janiskelemen/waf-admin/internal/api"
	"github.com/janiskelemen/waf-admin/internal/reload"
)

const geoIPFileName = "GeoLite2-Country.mmdb"

// RunGeoIPUpdate downloads the latest GeoLite2-Country database, writes it
// atomically to the configured directory, then stops Caddy so Docker's
// restart policy brings it back with the fresh database loaded.
func RunGeoIPUpdate(ctx context.Context, cfg api.GeoIPConfig, caddy *reload.CaddyAdmin) error {
	dest := filepath.Join(cfg.DatabaseDir, geoIPFileName)

	if err := downloadFile(ctx, cfg.DatabaseURL, dest); err != nil {
		log.Error().Err(err).Msg("geoip update: download failed")
		return err
	}
	log.Info().Str("path", dest).Msg("geoip update: database written")

	if err := caddy.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("geoip update: caddy stop failed")
		return err
	}
	log.Info().Msg("geoip update: caddy stopped, waiting for restart")
	return nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".geoip-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
