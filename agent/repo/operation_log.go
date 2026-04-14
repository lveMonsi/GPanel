package repo

import (
	"gpanel/agent/global"
	"gpanel/agent/models"
	"time"
)

type OperationLogRepo struct{}

func NewOperationLogRepo() *OperationLogRepo {
	return &OperationLogRepo{}
}

func (r *OperationLogRepo) Create(log *models.OperationLog) error {
	return global.DB.Create(log).Error
}

func (r *OperationLogRepo) Search(page, pageSize int, username, resource, action, status, keyword, startTime, endTime string) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	query := global.DB.Model(&models.OperationLog{})
	if username != "" {
		query = query.Where("username = ?", username)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("detail LIKE ?", "%"+keyword+"%")
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *OperationLogRepo) CleanBefore(days int) (int64, error) {
	before := time.Now().AddDate(0, 0, -days)
	result := global.DB.Where("created_at < ?", before).Delete(&models.OperationLog{})
	return result.RowsAffected, result.Error
}

func (r *OperationLogRepo) GetTotal() (int64, error) {
	var total int64
	err := global.DB.Model(&models.OperationLog{}).Count(&total).Error
	return total, err
}

func (r *OperationLogRepo) GetTodayCount() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := global.DB.Model(&models.OperationLog{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

func (r *OperationLogRepo) GetResourceStats() (map[string]int64, error) {
	type result struct {
		Resource string
		Count    int64
	}
	var results []result
	err := global.DB.Model(&models.OperationLog{}).
		Select("resource, count(*) as count").
		Group("resource").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Resource] = r.Count
	}
	return stats, nil
}
