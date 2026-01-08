package global

import (
	"context"
	"log"
	"sync"
	"time"
)

type ConfigReloader struct {
	mu             sync.Mutex
	cancelFunc     context.CancelFunc
	reloadInterval time.Duration
	onReload       []func()
}

var ConfigReloaderInstance *ConfigReloader

func InitConfigReloader(interval time.Duration) {
	ConfigReloaderInstance = &ConfigReloader{
		reloadInterval: interval,
		onReload:       make([]func(), 0),
	}

	if interval > 0 {
		ConfigReloaderInstance.Start()
		log.Printf("Config reloader initialized with interval: %v", interval)
	}
}

func (cr *ConfigReloader) Start() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.cancelFunc != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	cr.cancelFunc = cancel

	go cr.watch(ctx)
}

func (cr *ConfigReloader) Stop() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.cancelFunc != nil {
		cr.cancelFunc()
		cr.cancelFunc = nil
	}
}

func (cr *ConfigReloader) watch(ctx context.Context) {
	ticker := time.NewTicker(cr.reloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cr.reload()
		}
	}
}

func (cr *ConfigReloader) reload() error {
	if err := ConfigCacheInstance.Reload(); err != nil {
		log.Printf("Failed to reload config cache: %v", err)
		return err
	}

	for _, callback := range cr.onReload {
		if callback != nil {
			callback()
		}
	}

	log.Printf("Config reloaded successfully")
	return nil
}

func (cr *ConfigReloader) ReloadNow() error {
	return cr.reload()
}

func (cr *ConfigReloader) OnReload(callback func()) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.onReload = append(cr.onReload, callback)
}

func ReloadConfig() error {
	if ConfigReloaderInstance != nil {
		return ConfigReloaderInstance.reload()
	}
	return ConfigCacheInstance.Reload()
}