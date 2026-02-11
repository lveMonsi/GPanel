package service

import (
	"gpanel/dto"
	"gpanel/repo"
	"gpanel/models"
	"time"
)

type SessionHistoryService struct {
	repo repo.ISessionHistoryRepo
}

type ISessionHistoryService interface {
	Create(req dto.SessionHistoryCreate) (*models.SessionHistory, error)
	Update(req dto.SessionHistoryUpdate) error
	GetByID(id uint) (*dto.SessionHistoryInfo, error)
	List(page, pageSize int) ([]dto.SessionHistoryInfo, int64, error)
	Search(req dto.SessionHistorySearch) ([]dto.SessionHistoryInfo, int64, error)
	Delete(id uint) error
	DeleteByHostID(hostID uint) error
}

func NewSessionHistoryService() ISessionHistoryService {
	return &SessionHistoryService{
		repo: repo.NewSessionHistoryRepo(),
	}
}

func (s *SessionHistoryService) Create(req dto.SessionHistoryCreate) (*models.SessionHistory, error) {
	history := &models.SessionHistory{
		HostID:    req.HostID,
		HostAddr:  req.HostAddr,
		UserName:  req.UserName,
		StartTime: time.Now().Unix(),
		EndTime:   0,
		Duration:  0,
		Commands:  "[]",
	}
	if err := s.repo.Create(history); err != nil {
		return nil, err
	}
	return history, nil
}

func (s *SessionHistoryService) Update(req dto.SessionHistoryUpdate) error {
	history, err := s.repo.GetByID(req.ID)
	if err != nil {
		return err
	}
	history.EndTime = req.EndTime
	history.Duration = req.Duration
	history.Commands = req.Commands
	return s.repo.Update(history)
}

func (s *SessionHistoryService) GetByID(id uint) (*dto.SessionHistoryInfo, error) {
	history, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToDTO(history), nil
}

func (s *SessionHistoryService) List(page, pageSize int) ([]dto.SessionHistoryInfo, int64, error) {
	histories, total, err := s.repo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.SessionHistoryInfo
	for _, h := range histories {
		result = append(result, *s.convertToDTO(&h))
	}
	return result, total, nil
}

func (s *SessionHistoryService) Search(req dto.SessionHistorySearch) ([]dto.SessionHistoryInfo, int64, error) {
	histories, total, err := s.repo.Search(
		req.Page, req.PageSize,
		int(req.HostID), req.HostAddr, req.UserName,
		req.StartDate, req.EndDate,
	)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.SessionHistoryInfo
	for _, h := range histories {
		result = append(result, *s.convertToDTO(&h))
	}
	return result, total, nil
}

func (s *SessionHistoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *SessionHistoryService) DeleteByHostID(hostID uint) error {
	return s.repo.DeleteByHostID(hostID)
}

func (s *SessionHistoryService) convertToDTO(history *models.SessionHistory) *dto.SessionHistoryInfo {
	return &dto.SessionHistoryInfo{
		ID:        history.ID,
		CreatedAt: history.CreatedAt.Format(time.RFC3339),
		UpdatedAt: history.UpdatedAt.Format(time.RFC3339),
		HostID:    history.HostID,
		HostAddr:  history.HostAddr,
		UserName:  history.UserName,
		StartTime: history.StartTime,
		EndTime:   history.EndTime,
		Duration:  history.Duration,
		Commands:  history.Commands,
	}
}