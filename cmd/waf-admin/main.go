package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/janiskelemen/waf-admin/internal/api"
	"github.com/janiskelemen/waf-admin/internal/reload"
	"github.com/janiskelemen/waf-admin/internal/render"
	"github.com/janiskelemen/waf-admin/internal/scheduler"
	"github.com/janiskelemen/waf-admin/internal/storage"
	"github.com/janiskelemen/waf-admin/internal/util"
)

func main() {
	cfgPath := flag.String("config", "configs/config.example.yaml", "config file")
	flag.Parse()

	cfg, err := api.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	util.SetupLogging()

	stor := storage.NewFS()

	driver := render.NewCaddyCoraza(render.CaddyOptions{
		AdminSocket: cfg.Caddy.AdminSocket,
		Caddyfile:   cfg.Caddy.Caddyfile,
		SitesDir:    cfg.Caddy.SitesDir,
		RulesRoot:   cfg.Caddy.RulesRoot,
	})

	rl := reload.NewCaddyAdmin(cfg.Caddy.AdminSocket, cfg.Caddy.Caddyfile)

	sched := scheduler.New()
	if cfg.Backup.Enabled {
		if err := sched.AddDaily("backup", cfg.Backup.Daily, func(ctx context.Context) error {
			return scheduler.RunBackup(ctx, cfg.Backup, cfg.Caddy)
		}); err != nil {
			log.Fatal().Err(err).Msg("schedule backup")
		}
	}
	sched.Start()
	defer sched.Stop()

	srv := api.NewServer(cfg, stor, driver, rl)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("http server")
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
	<-sigC

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Stop(ctx)
}
