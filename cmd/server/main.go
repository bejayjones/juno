package main

import (
	"log/slog"
	"os"

	"github.com/bejayjones/juno/api/rest"
	"github.com/bejayjones/juno/pkg/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting juno",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
		"db_driver", cfg.Database.Driver,
	)

	srv := rest.NewServer(cfg)
	if err := srv.Start(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
