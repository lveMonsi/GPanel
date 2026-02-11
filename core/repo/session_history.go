package repo

import (
	"gpanel/global"
	"gpanel/models"
)

type SessionHistoryRepo struct{}

type ISessionHistoryRepo interface {
	Create(history *models.SessionHistory) error
	Update(history *models.SessionHistory) error
	GetByID(id uint) (*models.SessionHistory, error)
	List(page, pageSize int) ([]models.SessionHistory, int64, error)
	Search(page, pageSize, hostID int, hostAddr, userName string, startDate, endDate int64) ([]models.SessionHistory, int64, error)
	Delete(id uint) error
	DeleteByHostID(hostID uint) error
}

func NewSessionHistoryRepo() ISessionHistoryRepo {
	return &SessionHistoryRepo{}
}

func (r *SessionHistoryRepo) Create(history *models.SessionHistory) error {
	return global.DB.Create(history).Error
}

func (r *SessionHistoryRepo) Update(history *models.SessionHistory) error {
	return global.DB.Save(history).Error
}

func (r *SessionHistoryRepo) GetByID(id uint) (*models.SessionHistory, error) {
	var history models.SessionHistory
	err := global.DB.First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *SessionHistoryRepo) List(page, pageSize int) ([]models.SessionHistory, int64, error) {
	var histories []models.SessionHistory
	var total int64

	if err := global.DB.Model(&models.SessionHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := global.DB.Order("id DESC").Offset(offset).Limit(pageSize).Find(&histories).Error
	return histories, total, err
}

func (r *SessionHistoryRepo) Search(page, pageSize, hostID int, hostAddr, userName string, startDate, endDate int64) ([]models.SessionHistory, int64, error) {
	var histories []models.SessionHistory
	var total int64

	query := global.DB.Model(&models.SessionHistory{})

	if hostID > 0 {
		query = query.Where("host_id = ?", hostID)
	}
	if hostAddr != "" {
		query = query.Where("host_addr LIKE ?", "%"+hostAddr+"%")
	}
	if userName != "" {
		query = query.Where("user_name LIKE ?", "%"+userName+"%")
	}
	if startDate > 0 {
		query = query.Where("start_time >= ?", startDate)
	}
	if endDate > 0 {
		query = query.Where("end_time <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&histories).Error
	return histories, total, err
}

func (r *SessionHistoryRepo) Delete(id uint) error {
	return global.DB.Delete(&models.SessionHistory{}, id).Error
}

func (r *SessionHistoryRepo) DeleteByHostID(hostID uint) error {
	return global.DB.Where("host_id = ?", hostID).Delete(&models.SessionHistory{}).Error
}