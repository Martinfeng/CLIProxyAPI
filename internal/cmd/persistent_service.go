// Package cmd provides command-line interface functionality for the CLI Proxy API server.
// This file provides a persistence-enabled wrapper around the standard service startup.
package cmd

import (
	"context"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/usage"
	log "github.com/sirupsen/logrus"
)

// StartServiceWithPersistence is kept as a compatibility alias.
func StartServiceWithPersistence(cfg *config.Config, configPath string, localPassword string) {
	StartService(cfg, configPath, localPassword)
}

// StartServiceBackgroundWithPersistence is kept as a compatibility alias.
func StartServiceBackgroundWithPersistence(cfg *config.Config, configPath string, localPassword string) (cancel func(), done <-chan struct{}) {
	return StartServiceBackground(cfg, configPath, localPassword)
}

func startUsagePersistence(cfg *config.Config) {
	if cfg == nil || !cfg.UsageStatisticsEnabled || usage.IsPersistenceRunning() {
		return
	}
	if err := usage.StartPersistence(cfg.AuthDir); err != nil {
		log.Warnf("Failed to start usage persistence: %v", err)
	}
}

func stopUsagePersistence() {
	if !usage.IsPersistenceRunning() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := usage.StopPersistence(ctx); err != nil {
		log.Errorf("Failed to stop usage persistence: %v", err)
	}
}
