package repo

import (
	"gpanel/agent/global"
	"gpanel/agent/models"
)

type CronjobRepo struct{}

func NewCronjobRepo() *CronjobRepo {
	return &CronjobRepo{}
}

func (r *CronjobRepo) Create(job *models.Cronjob) error {
	return global.DB.Create(job).Error
}

func (r *CronjobRepo) Update(job *models.Cronjob) error {
	return global.DB.Save(job).Error
}

func (r *CronjobRepo) DeleteByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.Where("id IN ?", ids).Delete(&models.Cronjob{}).Error
}

func (r *CronjobRepo) GetByID(id uint) (*models.Cronjob, error) {
	var job models.Cronjob
	err := global.DB.First(&job, id).Error
	return &job, err
}

func (r *CronjobRepo) ListEnabled() ([]models.Cronjob, error) {
	var jobs []models.Cronjob
	err := global.DB.Where("status = ?", "enabled").Find(&jobs).Error
	return jobs, err
}

func (r *CronjobRepo) Search(page, pageSize int, taskType, status, keyword string) ([]models.Cronjob, int64, error) {
	var jobs []models.Cronjob
	var total int64

	query := global.DB.Model(&models.Cronjob{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// JobRecordRepo 执行记录仓库
type JobRecordRepo struct{}

func NewJobRecordRepo() *JobRecordRepo {
	return &JobRecordRepo{}
}

func (r *JobRecordRepo) Create(record *models.JobRecord) error {
	return global.DB.Create(record).Error
}

func (r *JobRecordRepo) Update(record *models.JobRecord) error {
	return global.DB.Save(record).Error
}

func (r *JobRecordRepo) GetByID(id uint) (*models.JobRecord, error) {
	var record models.JobRecord
	err := global.DB.First(&record, id).Error
	return &record, err
}

func (r *JobRecordRepo) GetLatestByCronjobID(cronjobID uint) (*models.JobRecord, error) {
	var record models.JobRecord
	err := global.DB.Where("cronjob_id = ?", cronjobID).Order("id DESC").First(&record).Error
	return &record, err
}

func (r *JobRecordRepo) Search(page, pageSize int, cronjobID uint, status string) ([]models.JobRecord, int64, error) {
	var records []models.JobRecord
	var total int64

	query := global.DB.Model(&models.JobRecord{}).Where("cronjob_id = ?", cronjobID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *JobRecordRepo) DeleteByCronjobID(cronjobID uint) error {
	return global.DB.Where("cronjob_id = ?", cronjobID).Delete(&models.JobRecord{}).Error
}
