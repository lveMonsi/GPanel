package service

import (
	"gpanel/agent/dto"
	"gpanel/agent/models"
	"gpanel/agent/repo"
)

type OperationLogService struct {
	logRepo *repo.OperationLogRepo
}

func NewOperationLogService() *OperationLogService {
	return &OperationLogService{
		logRepo: repo.NewOperationLogRepo(),
	}
}

func (s *OperationLogService) Create(req dto.OperationLogCreate) error {
	log := &models.OperationLog{
		Username: req.Username,
		IP:       req.IP,
		Resource: req.Resource,
		Action:   req.Action,
		Detail:   req.Detail,
		Status:   req.Status,
	}
	return s.logRepo.Create(log)
}

func (s *OperationLogService) Search(req dto.OperationLogSearch) ([]models.OperationLog, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	return s.logRepo.Search(
		req.Page, req.PageSize,
		req.Username, req.Resource, req.Action, req.Status,
		req.Keyword, req.StartTime, req.EndTime,
	)
}

func (s *OperationLogService) Clean(req dto.OperationLogClean) (int64, error) {
	return s.logRepo.CleanBefore(req.RetainDays)
}

func (s *OperationLogService) Stats() (*dto.OperationLogStats, error) {
	total, err := s.logRepo.GetTotal()
	if err != nil {
		return nil, err
	}
	todayCount, err := s.logRepo.GetTodayCount()
	if err != nil {
		return nil, err
	}
	resourceStat, err := s.logRepo.GetResourceStats()
	if err != nil {
		return nil, err
	}
	return &dto.OperationLogStats{
		Total:        total,
		TodayCount:   todayCount,
		ResourceStat: resourceStat,
	}, nil
}
