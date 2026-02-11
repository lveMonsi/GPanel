package service

import (
	"errors"
	"gpanel/dto"
	"gpanel/models"
	"gpanel/repo"
	"time"

	"gorm.io/gorm"
)

type QuickCommandService struct {
	quickCommandRepo repo.IQuickCommandRepo
}

type IQuickCommandService interface {
	Create(req dto.QuickCommandCreate) error
	Update(req dto.QuickCommandUpdate) error
	Delete(req dto.QuickCommandDelete) error
	Search(req dto.QuickCommandSearch) (*dto.QuickCommandPageResult, error)
	GetByID(id uint) (*dto.QuickCommandItem, error)
	GetAll() ([]dto.QuickCommandItem, error)
}

func NewQuickCommandService() IQuickCommandService {
	return &QuickCommandService{
		quickCommandRepo: repo.NewQuickCommandRepo(),
	}
}

// Create 创建快速命令
func (s *QuickCommandService) Create(req dto.QuickCommandCreate) error {
	cmd := &models.QuickCommand{
		Name:        req.Name,
		Command:     req.Command,
		Description: req.Description,
		GroupID:     req.GroupID,
		Sort:        req.Sort,
	}
	return s.quickCommandRepo.Create(cmd)
}

// Update 更新快速命令
func (s *QuickCommandService) Update(req dto.QuickCommandUpdate) error {
	cmd, err := s.quickCommandRepo.GetByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("quick command not found")
		}
		return err
	}

	// 更新字段
	if req.Name != "" {
		cmd.Name = req.Name
	}
	if req.Command != "" {
		cmd.Command = req.Command
	}
	if req.Description != "" {
		cmd.Description = req.Description
	}
	cmd.GroupID = req.GroupID
	cmd.Sort = req.Sort

	return s.quickCommandRepo.Update(cmd)
}

// Delete 删除快速命令
func (s *QuickCommandService) Delete(req dto.QuickCommandDelete) error {
	return s.quickCommandRepo.DeleteByIDs(req.IDs)
}

// Search 搜索快速命令
func (s *QuickCommandService) Search(req dto.QuickCommandSearch) (*dto.QuickCommandPageResult, error) {
	cmds, total, err := s.quickCommandRepo.Search(req.Page, req.PageSize, req.Keyword, req.GroupID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.QuickCommandItem, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, dto.QuickCommandItem{
			ID:          cmd.ID,
			Name:        cmd.Name,
			Command:     cmd.Command,
			Description: cmd.Description,
			GroupID:     cmd.GroupID,
			Sort:        cmd.Sort,
			CreatedAt:   cmd.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   cmd.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &dto.QuickCommandPageResult{
		Items: items,
		Total: total,
	}, nil
}

// GetByID 根据 ID 获取快速命令
func (s *QuickCommandService) GetByID(id uint) (*dto.QuickCommandItem, error) {
	cmd, err := s.quickCommandRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quick command not found")
		}
		return nil, err
	}

	return &dto.QuickCommandItem{
		ID:          cmd.ID,
		Name:        cmd.Name,
		Command:     cmd.Command,
		Description: cmd.Description,
		GroupID:     cmd.GroupID,
		Sort:        cmd.Sort,
		CreatedAt:   cmd.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   cmd.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetAll 获取所有快速命令
func (s *QuickCommandService) GetAll() ([]dto.QuickCommandItem, error) {
	cmds, err := s.quickCommandRepo.List()
	if err != nil {
		return nil, err
	}

	items := make([]dto.QuickCommandItem, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, dto.QuickCommandItem{
			ID:          cmd.ID,
			Name:        cmd.Name,
			Command:     cmd.Command,
			Description: cmd.Description,
			GroupID:     cmd.GroupID,
			Sort:        cmd.Sort,
			CreatedAt:   cmd.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   cmd.UpdatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}