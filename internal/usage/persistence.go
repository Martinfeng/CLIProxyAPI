// Package usage provides usage tracking and persistence functionality.
// This file implements automatic persistence of usage statistics to disk,
// enabling data to survive container restarts in Docker deployments.
package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Persistence manages automatic saving and loading of usage statistics.
type Persistence struct {
	path     string
	interval time.Duration

	running  atomic.Bool
	stopCh   chan struct{}
	doneCh   chan struct{}
	lastSave atomic.Int64

	saveMu sync.Mutex // Protects concurrent save operations
}

// persistencePayload matches the format used by /management/usage/export API.
type persistencePayload struct {
	Version   int                `json:"version"`
	SavedAt   time.Time          `json:"saved_at"`
	Usage     StatisticsSnapshot `json:"usage"`
	AutoSaved bool               `json:"auto_saved"`
}

// supportedVersions defines the payload versions this code can handle.
const (
	currentPayloadVersion = 1
	maxSupportedVersion   = 1
)

var (
	defaultPersistence *Persistence
	persistenceMu      sync.Mutex
)

// StartPersistence initializes and starts the persistence manager.
// It loads any existing statistics from disk and begins periodic auto-save.
//
// Parameters:
//   - authDir: The directory where usage-stats.json will be stored
//
// Returns:
//   - error: An error if initialization fails
func StartPersistence(authDir string) error {
	persistenceMu.Lock()
	defer persistenceMu.Unlock()

	if defaultPersistence != nil && defaultPersistence.running.Load() {
		return nil // Already running
	}

	// Determine storage path
	path := os.Getenv("USAGE_STATS_PATH")
	if path == "" {
		path = filepath.Join(authDir, "usage-stats.json")
	}

	// Determine auto-save interval
	interval := 5 * time.Minute
	if envInterval := os.Getenv("USAGE_STATS_AUTOSAVE_INTERVAL"); envInterval != "" {
		if parsed, err := time.ParseDuration(envInterval); err == nil && parsed > 0 {
			interval = parsed
		} else {
			log.Warnf("Invalid USAGE_STATS_AUTOSAVE_INTERVAL %q, using default 5m", envInterval)
		}
	}

	p := &Persistence{
		path:     path,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}

	// Load existing statistics
	if err := p.load(); err != nil {
		log.Warnf("Failed to load persisted usage statistics: %v", err)
		// Continue anyway - this is not fatal
	}

	// Start auto-save goroutine
	p.running.Store(true)
	go p.autoSaveLoop()

	defaultPersistence = p
	log.Infof("Usage statistics persistence enabled: path=%s, interval=%s", path, interval)
	return nil
}

// StopPersistence stops the persistence manager and performs a final save.
//
// Parameters:
//   - ctx: Context for controlling the shutdown timeout
//
// Returns:
//   - error: An error if the final save fails
func StopPersistence(ctx context.Context) error {
	persistenceMu.Lock()
	p := defaultPersistence
	persistenceMu.Unlock()

	if p == nil || !p.running.Load() {
		return nil
	}

	return p.stop(ctx)
}

// SaveNow triggers an immediate save of the current statistics.
// This can be called manually if needed.
//
// Returns:
//   - error: An error if the save fails
func SaveNow() error {
	persistenceMu.Lock()
	p := defaultPersistence
	persistenceMu.Unlock()

	if p == nil {
		return fmt.Errorf("persistence not initialized")
	}
	return p.save()
}

// load reads statistics from disk and merges them into memory.
func (p *Persistence) load() error {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug("No existing usage statistics file found, starting fresh")
			return nil
		}
		// Try backup file
		backupPath := p.path + ".bak"
		data, err = os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read usage statistics: %w", err)
		}
		log.Warn("Loaded usage statistics from backup file")
	}

	var payload persistencePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		// Try backup file
		backupPath := p.path + ".bak"
		backupData, backupErr := os.ReadFile(backupPath)
		if backupErr != nil {
			return fmt.Errorf("failed to parse usage statistics: %w", err)
		}
		if err := json.Unmarshal(backupData, &payload); err != nil {
			return fmt.Errorf("failed to parse usage statistics from backup: %w", err)
		}
		log.Warn("Loaded usage statistics from backup file due to corrupted main file")
	}

	// Validate version
	if payload.Version > maxSupportedVersion {
		log.Warnf("Usage statistics file has unsupported version %d (max supported: %d), data may be incompatible",
			payload.Version, maxSupportedVersion)
	}

	// Merge into current statistics
	stats := GetRequestStatistics()
	if stats != nil {
		result := stats.MergeSnapshot(payload.Usage)
		stats.PruneRetention(time.Now())
		log.Infof("Loaded usage statistics: added=%d, skipped=%d (duplicates)", result.Added, result.Skipped)
	}

	return nil
}

// save writes the current statistics to disk atomically.
// This method is safe for concurrent calls.
func (p *Persistence) save() error {
	p.saveMu.Lock()
	defer p.saveMu.Unlock()

	stats := GetRequestStatistics()
	if stats == nil {
		return nil
	}

	// Prune old data before saving
	stats.PruneRetention(time.Now())

	snapshot := stats.Snapshot()
	payload := persistencePayload{
		Version:   currentPayloadVersion,
		SavedAt:   time.Now().UTC(),
		Usage:     snapshot,
		AutoSaved: true,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal usage statistics: %w", err)
	}

	// Ensure directory exists (using restrictive permissions for auth-related data)
	dir := filepath.Dir(p.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Atomic write: write to temp file, fsync, then rename
	tmpFile, err := os.CreateTemp(dir, "usage-stats-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write data
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to ensure data is on disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Create backup of existing file
	if _, err := os.Stat(p.path); err == nil {
		backupPath := p.path + ".bak"
		// Remove existing backup first to avoid rename issues on some platforms
		if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
			log.Debugf("Failed to remove old backup file: %v", err)
		}
		if err := os.Rename(p.path, backupPath); err != nil {
			log.Warnf("Failed to create backup of usage statistics: %v", err)
			// Continue anyway - backup is not critical
		}
	}

	// Rename temp to final
	if err := os.Rename(tmpPath, p.path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	p.lastSave.Store(time.Now().Unix())
	log.Debugf("Saved usage statistics: requests=%d, tokens=%d",
		snapshot.TotalRequests, snapshot.TotalTokens)
	return nil
}

// autoSaveLoop periodically saves statistics to disk.
func (p *Persistence) autoSaveLoop() {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.save(); err != nil {
				log.Errorf("Auto-save usage statistics failed: %v", err)
			}
		case <-p.stopCh:
			return
		}
	}
}

// stop gracefully stops the persistence manager and performs a final save.
func (p *Persistence) stop(ctx context.Context) error {
	if !p.running.CompareAndSwap(true, false) {
		return nil // Already stopped
	}

	close(p.stopCh)

	// Wait for auto-save loop to finish
	select {
	case <-p.doneCh:
	case <-ctx.Done():
		log.Warn("Timeout waiting for auto-save loop to stop")
	}

	// Final save
	log.Info("Performing final usage statistics save...")
	if err := p.save(); err != nil {
		log.Errorf("Final save failed: %v", err)
		return err
	}
	log.Info("Usage statistics saved successfully")
	return nil
}

// GetPersistencePath returns the current persistence file path, or empty if not initialized.
func GetPersistencePath() string {
	persistenceMu.Lock()
	defer persistenceMu.Unlock()
	if defaultPersistence == nil {
		return ""
	}
	return defaultPersistence.path
}

// IsPersistenceRunning returns whether persistence is currently active.
func IsPersistenceRunning() bool {
	persistenceMu.Lock()
	defer persistenceMu.Unlock()
	return defaultPersistence != nil && defaultPersistence.running.Load()
}
