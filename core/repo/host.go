package repo

import (
	"gpanel/global"
	"gpanel/models"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	// HostGroup operations
	CreateGroup(group *models.HostGroup) error
	UpdateGroup(group *models.HostGroup) error
	DeleteGroup(id uint) error
	GetGroupByID(id uint) (*models.HostGroup, error)
	ListGroups() ([]models.HostGroup, error)
	SearchGroups(page, pageSize int, info string) ([]models.HostGroup, int64, error)
	GetGroupHostCount(id uint) (int64, error)

	// Host operations
	CreateHost(host *models.Host) error
	UpdateHost(host *models.Host) error
	DeleteHost(id uint) error
	GetHostByID(id uint) (*models.Host, error)
	ListHosts() ([]models.Host, error)
	SearchHosts(page, pageSize, groupID int, info string) ([]models.Host, int64, error)
	GetHostsByGroupID(groupID uint) ([]models.Host, error)
	MoveHostsToGroup(hostIDs []uint, groupID uint) error
}

func NewHostRepo() IHostRepo {
	return &HostRepo{}
}

// HostGroup operations

func (r *HostRepo) CreateGroup(group *models.HostGroup) error {
	return global.DB.Create(group).Error
}

func (r *HostRepo) UpdateGroup(group *models.HostGroup) error {
	return global.DB.Save(group).Error
}

func (r *HostRepo) DeleteGroup(id uint) error {
	// 先检查分组下是否有主机
	var count int64
	if err := global.DB.Model(&models.Host{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return gorm.ErrRecordNotFound // 分组下有主机，不能删除
	}
	return global.DB.Delete(&models.HostGroup{}, id).Error
}

func (r *HostRepo) GetGroupByID(id uint) (*models.HostGroup, error) {
	var group models.HostGroup
	err := global.DB.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *HostRepo) ListGroups() ([]models.HostGroup, error) {
	var groups []models.HostGroup
	err := global.DB.Find(&groups).Error
	return groups, err
}

func (r *HostRepo) SearchGroups(page, pageSize int, info string) ([]models.HostGroup, int64, error) {
	var groups []models.HostGroup
	var total int64

	query := global.DB.Model(&models.HostGroup{})
	if info != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+info+"%", "%"+info+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&groups).Error
	return groups, total, err
}

func (r *HostRepo) GetGroupHostCount(id uint) (int64, error) {
	var count int64
	err := global.DB.Model(&models.Host{}).Where("group_id = ?", id).Count(&count).Error
	return count, err
}

// Host operations

func (r *HostRepo) CreateHost(host *models.Host) error {
	return global.DB.Create(host).Error
}

func (r *HostRepo) UpdateHost(host *models.Host) error {
	return global.DB.Save(host).Error
}

func (r *HostRepo) DeleteHost(id uint) error {
	return global.DB.Delete(&models.Host{}, id).Error
}

func (r *HostRepo) GetHostByID(id uint) (*models.Host, error) {
	var host models.Host
	err := global.DB.Preload("HostGroup").First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepo) ListHosts() ([]models.Host, error) {
	var hosts []models.Host
	err := global.DB.Preload("HostGroup").Find(&hosts).Error
	return hosts, err
}

func (r *HostRepo) SearchHosts(page, pageSize, groupID int, info string) ([]models.Host, int64, error) {
	var hosts []models.Host
	var total int64

	query := global.DB.Model(&models.Host{})

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}

	if info != "" {
		query = query.Where("name LIKE ? OR addr LIKE ? OR user LIKE ?", "%"+info+"%", "%"+info+"%", "%"+info+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("HostGroup").Order("id ASC").Offset(offset).Limit(pageSize).Find(&hosts).Error
	return hosts, total, err
}

func (r *HostRepo) GetHostsByGroupID(groupID uint) ([]models.Host, error) {
	var hosts []models.Host
	err := global.DB.Where("group_id = ?", groupID).Find(&hosts).Error
	return hosts, err
}

func (r *HostRepo) MoveHostsToGroup(hostIDs []uint, groupID uint) error {
	return global.DB.Model(&models.Host{}).Where("id IN ?", hostIDs).Update("group_id", groupID).Error
}