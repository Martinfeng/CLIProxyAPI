// Package cmd provides command-line interface functionality for the CLI Proxy API server.
// This file provides a persistence-enabled wrapper around the standard service startup.
package cmd

import (
	"context"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/usage"
	log "github.com/sirupsen/logrus"
)

// StartServiceWithPersistence wraps StartService with automatic usage statistics persistence.
// It loads persisted statistics on startup and saves them on shutdown, ensuring data
// survives container restarts in Docker deployments.
//
// Parameters:
//   - cfg: The application configuration
//   - configPath: The path to the configuration file
//   - localPassword: Optional password accepted for local management requests
func StartServiceWithPersistence(cfg *config.Config, configPath string, localPassword string) {
	// Start persistence if statistics are enabled
	if cfg != nil && cfg.UsageStatisticsEnabled {
		if err := usage.StartPersistence(cfg.AuthDir); err != nil {
			log.Warnf("Failed to start usage persistence: %v", err)
			// Continue anyway - persistence is not critical
		}
	}

	// Run the original service
	StartService(cfg, configPath, localPassword)

	// Stop persistence after service exits
	if usage.IsPersistenceRunning() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := usage.StopPersistence(ctx); err != nil {
			log.Errorf("Failed to stop usage persistence: %v", err)
		}
	}
}
