package repo

import (
	"gpanel/global"
	"gpanel/models"
)

type QuickCommandRepo struct{}

type IQuickCommandRepo interface {
	Create(cmd *models.QuickCommand) error
	Update(cmd *models.QuickCommand) error
	Delete(id uint) error
	DeleteByIDs(ids []uint) error
	GetByID(id uint) (*models.QuickCommand, error)
	List() ([]models.QuickCommand, error)
	Search(page, pageSize int, keyword string, groupID uint) ([]models.QuickCommand, int64, error)
}

func NewQuickCommandRepo() IQuickCommandRepo {
	return &QuickCommandRepo{}
}

func (r *QuickCommandRepo) Create(cmd *models.QuickCommand) error {
	return global.DB.Create(cmd).Error
}

func (r *QuickCommandRepo) Update(cmd *models.QuickCommand) error {
	return global.DB.Save(cmd).Error
}

func (r *QuickCommandRepo) Delete(id uint) error {
	return global.DB.Delete(&models.QuickCommand{}, id).Error
}

func (r *QuickCommandRepo) DeleteByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.Where("id IN ?", ids).Delete(&models.QuickCommand{}).Error
}

func (r *QuickCommandRepo) GetByID(id uint) (*models.QuickCommand, error) {
	var cmd models.QuickCommand
	err := global.DB.First(&cmd, id).Error
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (r *QuickCommandRepo) List() ([]models.QuickCommand, error) {
	var cmds []models.QuickCommand
	err := global.DB.Order("sort ASC, id DESC").Find(&cmds).Error
	return cmds, err
}

func (r *QuickCommandRepo) Search(page, pageSize int, keyword string, groupID uint) ([]models.QuickCommand, int64, error) {
	var cmds []models.QuickCommand
	var total int64

	query := global.DB.Model(&models.QuickCommand{})

	if keyword != "" {
		keywordLike := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR command LIKE ? OR description LIKE ?", keywordLike, keywordLike, keywordLike)
	}

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&cmds).Error; err != nil {
		return nil, 0, err
	}

	return cmds, total, nil
}