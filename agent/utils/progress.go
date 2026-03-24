package utils

import (
	"fmt"
	"sync"
	"time"

	"gpanel/agent/dto"
)

type ProgressManager struct {
	mu       sync.RWMutex
	progress map[string]*dto.ProgressInfo
}

var globalProgressManager *ProgressManager
var progressManagerOnce sync.Once

func GetProgressManager() *ProgressManager {
	progressManagerOnce.Do(func() {
		globalProgressManager = &ProgressManager{progress: make(map[string]*dto.ProgressInfo)}
	})
	return globalProgressManager
}

func (pm *ProgressManager) CreateProgress(key, name string, total int64) *dto.ProgressInfo {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	progress := &dto.ProgressInfo{Key: key, Name: name, Total: total, Status: "pending"}
	pm.progress[key] = progress
	return progress
}

func (pm *ProgressManager) UpdateProgress(key string, current int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if progress, ok := pm.progress[key]; ok {
		progress.Current = current
		if progress.Total > 0 {
			progress.Percent = float64(current) / float64(progress.Total) * 100
		}
	}
}

func (pm *ProgressManager) SetProgressStatus(key, status string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if progress, ok := pm.progress[key]; ok {
		progress.Status = status
	}
}

func (pm *ProgressManager) SetProgressMessage(key, message string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if progress, ok := pm.progress[key]; ok {
		progress.Message = message
	}
}

func (pm *ProgressManager) GetProgress(key string) *dto.ProgressInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.progress[key]
}

func GenerateProgressKey(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
