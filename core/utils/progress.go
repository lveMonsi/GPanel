package utils

import (
	"fmt"
	"sync"
	"time"

	"gpanel/dto"
)

// ProgressManager 进度管理器
type ProgressManager struct {
	mu       sync.RWMutex
	progress map[string]*dto.ProgressInfo
}

var globalProgressManager *ProgressManager
var progressManagerOnce sync.Once

// GetProgressManager 获取全局进度管理器
func GetProgressManager() *ProgressManager {
	progressManagerOnce.Do(func() {
		globalProgressManager = &ProgressManager{
			progress: make(map[string]*dto.ProgressInfo),
		}
	})
	return globalProgressManager
}

// CreateProgress 创建新的进度
func (pm *ProgressManager) CreateProgress(key, name string, total int64) *dto.ProgressInfo {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	progress := &dto.ProgressInfo{
		Key:     key,
		Name:    name,
		Total:   total,
		Current: 0,
		Percent: 0,
		Status:  "pending",
		Message: "",
	}

	pm.progress[key] = progress
	return progress
}

// UpdateProgress 更新进度
func (pm *ProgressManager) UpdateProgress(key string, current int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if progress, ok := pm.progress[key]; ok {
		progress.Current = current
		if progress.Total > 0 {
			progress.Percent = float64(current) / float64(progress.Total) * 100
		}
		
		// 通过 WebSocket 推送进度更新
		BroadcastProgress(key, progress)
	}
}

// SetProgressStatus 设置进度状态
func (pm *ProgressManager) SetProgressStatus(key, status string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if progress, ok := pm.progress[key]; ok {
		progress.Status = status
	}
}

// SetProgressMessage 设置进度消息
func (pm *ProgressManager) SetProgressMessage(key, message string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if progress, ok := pm.progress[key]; ok {
		progress.Message = message
	}
}

// GetProgress 获取进度
func (pm *ProgressManager) GetProgress(key string) *dto.ProgressInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if progress, ok := pm.progress[key]; ok {
		return progress
	}
	return nil
}

// DeleteProgress 删除进度
func (pm *ProgressManager) DeleteProgress(key string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.progress, key)
}

// CleanupOldProgress 清理旧的进度记录（超过5分钟）
func (pm *ProgressManager) CleanupOldProgress() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for key, progress := range pm.progress {
		if progress.Status == "completed" || progress.Status == "failed" {
			// 假设5分钟后清理
			if time.Since(now) > 5*time.Minute {
				delete(pm.progress, key)
			}
		}
	}
}

// GenerateProgressKey 生成进度键
func (pm *ProgressManager) GenerateProgressKey() string {
	return fmt.Sprintf("progress_%d", time.Now().UnixNano())
}

// StartCleanupTask 启动清理任务
func StartCleanupTask() {
	pm := GetProgressManager()
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			pm.CleanupOldProgress()
		}
	}()
}
func GenerateProgressKey(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}