package service

import (
	"context"
	"errors"
	"fmt"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/models"
	"gpanel/agent/repo"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type CronjobService struct {
	cronjobRepo  *repo.CronjobRepo
	recordRepo   *repo.JobRecordRepo
	runningTasks sync.Map
}

func NewCronjobService() *CronjobService {
	return &CronjobService{
		cronjobRepo: repo.NewCronjobRepo(),
		recordRepo:  repo.NewJobRecordRepo(),
	}
}

func (s *CronjobService) Create(req dto.CronjobCreate) error {
	job := &models.Cronjob{
		Name:           req.Name,
		Type:           req.Type,
		Spec:           req.Spec,
		SpecCustom:     req.SpecCustom,
		Script:         req.Script,
		URL:            req.URL,
		SourceDir:      req.SourceDir,
		ExclusionRules: req.ExclusionRules,
		RetainCopies:   req.RetainCopies,
		RetryCount:     req.RetryCount,
		Timeout:        req.Timeout,
		IgnoreErr:      req.IgnoreErr,
		Status:         "enabled",
	}
	if err := s.cronjobRepo.Create(job); err != nil {
		return err
	}
	return s.registerCronEntry(job)
}

func (s *CronjobService) Update(req dto.CronjobUpdate) error {
	job, err := s.cronjobRepo.GetByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task not found")
		}
		return err
	}

	s.removeCronEntry(job)

	if req.Name != "" {
		job.Name = req.Name
	}
	if req.Spec != "" {
		job.Spec = req.Spec
	}
	job.SpecCustom = req.SpecCustom
	job.Script = req.Script
	job.URL = req.URL
	job.SourceDir = req.SourceDir
	job.ExclusionRules = req.ExclusionRules
	job.RetainCopies = req.RetainCopies
	job.RetryCount = req.RetryCount
	job.Timeout = req.Timeout
	job.IgnoreErr = req.IgnoreErr

	if err := s.cronjobRepo.Update(job); err != nil {
		return err
	}
	if job.Status == "enabled" {
		return s.registerCronEntry(job)
	}
	return nil
}

func (s *CronjobService) Delete(req dto.CronjobDelete) error {
	for _, id := range req.IDs {
		job, err := s.cronjobRepo.GetByID(id)
		if err != nil {
			continue
		}
		s.removeCronEntry(job)
		s.recordRepo.DeleteByCronjobID(id)
		s.cleanLogDir(id)
	}
	return s.cronjobRepo.DeleteByIDs(req.IDs)
}

type cronjobItem struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Spec             string `json:"spec"`
	SpecCustom       bool   `json:"specCustom"`
	Script           string `json:"script"`
	URL              string `json:"url"`
	SourceDir        string `json:"sourceDir"`
	ExclusionRules   string `json:"exclusionRules"`
	RetainCopies     int    `json:"retainCopies"`
	RetryCount       int    `json:"retryCount"`
	Timeout          int    `json:"timeout"`
	IgnoreErr        bool   `json:"ignoreErr"`
	Status           string `json:"status"`
	LastRecordStatus string `json:"lastRecordStatus"`
	LastRecordTime   string `json:"lastRecordTime"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type cronjobPageResult struct {
	Items []cronjobItem `json:"items"`
	Total int64         `json:"total"`
}

func (s *CronjobService) Search(req dto.CronjobSearch) (*cronjobPageResult, error) {
	jobs, total, err := s.cronjobRepo.Search(req.Page, req.PageSize, req.Type, req.Status, req.Keyword)
	if err != nil {
		return nil, err
	}

	items := make([]cronjobItem, 0, len(jobs))
	for _, job := range jobs {
		item := cronjobItem{
			ID:             job.ID,
			Name:           job.Name,
			Type:           job.Type,
			Spec:           job.Spec,
			SpecCustom:     job.SpecCustom,
			Script:         job.Script,
			URL:            job.URL,
			SourceDir:      job.SourceDir,
			ExclusionRules: job.ExclusionRules,
			RetainCopies:   job.RetainCopies,
			RetryCount:     job.RetryCount,
			Timeout:        job.Timeout,
			IgnoreErr:      job.IgnoreErr,
			Status:         job.Status,
			CreatedAt:      job.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      job.UpdatedAt.Format(time.RFC3339),
		}

		record, err := s.recordRepo.GetLatestByCronjobID(job.ID)
		if err == nil {
			item.LastRecordStatus = record.Status
			item.LastRecordTime = time.UnixMilli(record.StartTime).Format("2006-01-02 15:04:05")
		}
		items = append(items, item)
	}

	return &cronjobPageResult{Items: items, Total: total}, nil
}

func (s *CronjobService) Toggle(req dto.CronjobToggle) error {
	job, err := s.cronjobRepo.GetByID(req.ID)
	if err != nil {
		return err
	}

	s.removeCronEntry(job)
	job.Status = req.Status

	if err := s.cronjobRepo.Update(job); err != nil {
		return err
	}
	if req.Status == "enabled" {
		return s.registerCronEntry(job)
	}
	return nil
}

func (s *CronjobService) HandleOnce(req dto.CronjobHandle) error {
	job, err := s.cronjobRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	go s.executeJob(job)
	return nil
}

func (s *CronjobService) StopRunning(req dto.CronjobStop) error {
	if cancel, ok := s.runningTasks.Load(req.ID); ok {
		cancel.(context.CancelFunc)()
		return nil
	}
	return errors.New("no running task found")
}

type recordItem struct {
	ID        uint   `json:"id"`
	CronjobID uint   `json:"cronjobId"`
	StartTime string `json:"startTime"`
	Duration  int64  `json:"duration"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type recordPageResult struct {
	Items []recordItem `json:"items"`
	Total int64        `json:"total"`
}

func (s *CronjobService) SearchRecords(req dto.RecordSearch) (*recordPageResult, error) {
	records, total, err := s.recordRepo.Search(req.Page, req.PageSize, req.CronjobID, req.Status)
	if err != nil {
		return nil, err
	}

	items := make([]recordItem, 0, len(records))
	for _, r := range records {
		items = append(items, recordItem{
			ID:        r.ID,
			CronjobID: r.CronjobID,
			StartTime: time.UnixMilli(r.StartTime).Format("2006-01-02 15:04:05"),
			Duration:  r.Duration,
			Status:    r.Status,
			Message:   r.Message,
		})
	}
	return &recordPageResult{Items: items, Total: total}, nil
}

func (s *CronjobService) GetRecordLog(recordID uint) (string, error) {
	record, err := s.recordRepo.GetByID(recordID)
	if err != nil {
		return "", err
	}
	if record.LogFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(record.LogFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *CronjobService) CleanRecords(req dto.RecordClean) error {
	s.cleanLogDir(req.CronjobID)
	return s.recordRepo.DeleteByCronjobID(req.CronjobID)
}

func (s *CronjobService) GetNextExecTimes(req dto.NextTimesReq) ([]string, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.Spec)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	times := make([]string, 0, 5)
	next := time.Now()
	for range 5 {
		next = schedule.Next(next)
		times = append(times, next.Format("2006-01-02 15:04:05"))
	}
	return times, nil
}

func (s *CronjobService) StartAllEnabled() error {
	jobs, err := s.cronjobRepo.ListEnabled()
	if err != nil {
		return err
	}
	for i := range jobs {
		if err := s.registerCronEntry(&jobs[i]); err != nil {
			log.Printf("failed to register cronjob [%s]: %v", jobs[i].Name, err)
		}
	}
	return nil
}

// registerCronEntry 注册 cron 调度条目
func (s *CronjobService) registerCronEntry(job *models.Cronjob) error {
	entryID, err := global.Cron.AddFunc(job.Spec, func() {
		j, err := s.cronjobRepo.GetByID(job.ID)
		if err != nil {
			return
		}
		s.executeJob(j)
	})
	if err != nil {
		return fmt.Errorf("failed to register cron: %w", err)
	}
	job.EntryID = int(entryID)
	return s.cronjobRepo.Update(job)
}

// removeCronEntry 移除 cron 调度条目
func (s *CronjobService) removeCronEntry(job *models.Cronjob) {
	if job.EntryID > 0 {
		global.Cron.Remove(cron.EntryID(job.EntryID))
		job.EntryID = 0
	}
}

// executeJob 执行计划任务
func (s *CronjobService) executeJob(job *models.Cronjob) {
	startTime := time.Now()

	record := &models.JobRecord{
		CronjobID: job.ID,
		StartTime: startTime.UnixMilli(),
		Status:    "waiting",
	}
	if err := s.recordRepo.Create(record); err != nil {
		log.Printf("failed to create job record: %v", err)
		return
	}

	logDir := filepath.Join(global.DataDirPath, "cronjob_logs", fmt.Sprintf("%d", job.ID))
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("%d.log", record.ID))
	logFile, err := os.Create(logPath)
	if err != nil {
		log.Printf("failed to create log file: %v", err)
		return
	}
	record.LogFile = logPath

	var ctx context.Context
	var cancel context.CancelFunc
	if job.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(job.Timeout)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	s.runningTasks.Store(job.ID, cancel)
	defer func() {
		cancel()
		s.runningTasks.Delete(job.ID)
		logFile.Close()
	}()

	var execErr error
	retries := job.RetryCount
	for attempt := range retries + 1 {
		if attempt > 0 {
			fmt.Fprintf(logFile, "\n--- retry %d/%d ---\n", attempt, retries)
		}
		switch job.Type {
		case "shell":
			execErr = s.executeShell(ctx, job, logFile)
		case "curl":
			execErr = s.executeCurl(ctx, job, logFile)
		case "directory":
			execErr = s.executeDirectory(ctx, job, logFile)
		case "clean":
			execErr = s.executeClean(ctx, logFile)
		case "cleanLog":
			execErr = s.executeCleanLog(ctx, logFile)
		default:
			execErr = fmt.Errorf("unsupported task type: %s", job.Type)
		}
		if execErr == nil {
			break
		}
	}

	record.Duration = time.Since(startTime).Milliseconds()
	if execErr != nil {
		if job.IgnoreErr {
			record.Status = "success"
			record.Message = fmt.Sprintf("completed with ignored error: %s", execErr.Error())
		} else {
			record.Status = "failed"
			record.Message = execErr.Error()
		}
	} else {
		record.Status = "success"
	}
	s.recordRepo.Update(record)
}

func (s *CronjobService) executeShell(ctx context.Context, job *models.Cronjob, w io.Writer) error {
	tmpFile, err := os.CreateTemp("", "gpanel-script-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(job.Script); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, "bash", tmpFile.Name())
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func (s *CronjobService) executeCurl(ctx context.Context, job *models.Cronjob, w io.Writer) error {
	urls := strings.Split(job.URL, "\n")
	client := &http.Client{Timeout: 30 * time.Second}

	for _, rawURL := range urls {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			fmt.Fprintf(w, "[ERROR] %s: %s\n", rawURL, err.Error())
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(w, "[ERROR] %s: %s\n", rawURL, err.Error())
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Fprintf(w, "[%d] %s\n%s\n\n", resp.StatusCode, rawURL, string(body))
		if resp.StatusCode >= 400 {
			return fmt.Errorf("request to %s returned status %d", rawURL, resp.StatusCode)
		}
	}
	return nil
}

func (s *CronjobService) executeDirectory(ctx context.Context, job *models.Cronjob, w io.Writer) error {
	if job.SourceDir == "" {
		return errors.New("source directory not specified")
	}

	backupDir := filepath.Join(global.DataDirPath, "cronjob_logs", fmt.Sprintf("%d", job.ID), "backups")
	os.MkdirAll(backupDir, 0755)

	archiveName := fmt.Sprintf("backup_%s.tar.gz", time.Now().Format("20060102_150405"))
	archivePath := filepath.Join(backupDir, archiveName)

	args := []string{"czf", archivePath}
	if job.ExclusionRules != "" {
		for _, rule := range strings.Split(job.ExclusionRules, "\n") {
			rule = strings.TrimSpace(rule)
			if rule != "" {
				args = append(args, "--exclude="+rule)
			}
		}
	}
	args = append(args, "-C", filepath.Dir(job.SourceDir), filepath.Base(job.SourceDir))

	cmd := exec.CommandContext(ctx, "tar", args...)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Fprintf(w, "backup saved to: %s\n", archivePath)

	if job.RetainCopies > 0 {
		s.cleanOldBackups(backupDir, job.RetainCopies, w)
	}
	return nil
}

func (s *CronjobService) executeClean(ctx context.Context, w io.Writer) error {
	dirs := []string{"/tmp", "/var/tmp"}
	var totalCleaned int64

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(w, "failed to read %s: %s\n", dir, err.Error())
			continue
		}
		cutoff := time.Now().Add(-24 * time.Hour)
		for _, entry := range entries {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				fullPath := filepath.Join(dir, entry.Name())
				size := info.Size()
				if entry.IsDir() {
					os.RemoveAll(fullPath)
				} else {
					os.Remove(fullPath)
				}
				totalCleaned += size
				fmt.Fprintf(w, "removed: %s (%d bytes)\n", fullPath, size)
			}
		}
	}
	fmt.Fprintf(w, "\ntotal cleaned: %d bytes\n", totalCleaned)
	return nil
}

func (s *CronjobService) executeCleanLog(ctx context.Context, w io.Writer) error {
	logDirs := []string{"/var/log"}
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var totalCleaned int64

	for _, dir := range logDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".log.1") || strings.HasSuffix(path, ".gz") {
				if info.ModTime().Before(cutoff) && info.Size() > 0 {
					size := info.Size()
					if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".log.1") {
						os.Remove(path)
						fmt.Fprintf(w, "removed: %s (%d bytes)\n", path, size)
					} else {
						os.Truncate(path, 0)
						fmt.Fprintf(w, "truncated: %s (%d bytes freed)\n", path, size)
					}
					totalCleaned += size
				}
			}
			return nil
		})
	}
	fmt.Fprintf(w, "\ntotal cleaned: %d bytes\n", totalCleaned)
	return nil
}

func (s *CronjobService) cleanOldBackups(dir string, retain int, w io.Writer) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type fileEntry struct {
		name    string
		modTime time.Time
	}
	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEntry{name: e.Name(), modTime: info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
	for i := retain; i < len(files); i++ {
		fullPath := filepath.Join(dir, files[i].name)
		os.Remove(fullPath)
		fmt.Fprintf(w, "removed old backup: %s\n", fullPath)
	}
}

func (s *CronjobService) cleanLogDir(cronjobID uint) {
	logDir := filepath.Join(global.DataDirPath, "cronjob_logs", fmt.Sprintf("%d", cronjobID))
	os.RemoveAll(logDir)
}
