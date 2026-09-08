package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type UpdateRequest struct {
	Version     string `json:"version"`
	Prerelease  bool   `json:"prerelease"`
	Accelerated bool   `json:"accelerated"`
	Force       bool   `json:"force"`
}

type UpdateStatus struct {
	State         string    `json:"state"`
	Phase         string    `json:"phase"`
	Version       string    `json:"version,omitempty"`
	TargetVersion string    `json:"target_version,omitempty"`
	Message       string    `json:"message,omitempty"`
	Percent       int       `json:"percent"`
	StartedAt     time.Time `json:"startedAt,omitempty"`
	FinishedAt    time.Time `json:"finishedAt,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type onlineStatus struct {
	Phase         string    `json:"phase"`
	Message       string    `json:"message"`
	Version       string    `json:"version,omitempty"`
	TargetVersion string    `json:"target_version,omitempty"`
	Percent       int       `json:"percent"`
	Error         string    `json:"error,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateService struct {
	mu      sync.Mutex
	running bool
}

func NewUpdateService() *UpdateService { return &UpdateService{} }

func updateStatusPath() string {
	if path := strings.TrimSpace(os.Getenv("GPANEL_UPDATE_STATUS_FILE")); path != "" {
		return path
	}
	return "/var/lib/gpanel/online-update/status.json"
}

func (s *UpdateService) Status() UpdateStatus {
	body, err := os.ReadFile(updateStatusPath())
	if err != nil {
		return UpdateStatus{State: "idle", Phase: "idle"}
	}
	var raw onlineStatus
	if err := json.Unmarshal(body, &raw); err != nil {
		return UpdateStatus{State: "failed", Phase: "failed", Error: "invalid update status"}
	}
	state := "idle"
	switch raw.Phase {
	case "checking", "downloading", "validating", "restarting", "rolling_back":
		state = "running"
	case "staged", "completed":
		state = "success"
	case "failed":
		state = "failed"
	}
	return UpdateStatus{
		State: state, Phase: raw.Phase, Version: raw.Version,
		TargetVersion: raw.TargetVersion, Message: raw.Message,
		Percent: raw.Percent, UpdatedAt: raw.UpdatedAt,
		FinishedAt: raw.UpdatedAt, Error: raw.Error,
	}
}

func validateUpdateRequest(req UpdateRequest) error {
	if len(req.Version) > 128 || strings.ContainsAny(req.Version, "\r\n/\\") {
		return errors.New("invalid version")
	}
	return nil
}

func gpctlPath() string {
	if path := strings.TrimSpace(os.Getenv("GPANEL_GPCTL_PATH")); path != "" {
		return path
	}
	return "/usr/local/bin/gpctl"
}

func (s *UpdateService) start(args []string) error {
	cmd := exec.Command(gpctlPath(), args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
	return nil
}

func (s *UpdateService) Stage(req UpdateRequest) error {
	if err := validateUpdateRequest(req); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.Status()
	if s.running || status.State == "running" {
		return errors.New("update already running")
	}
	args := []string{"online-stage"}
	if req.Version != "" {
		args = append(args, req.Version)
	}
	if req.Prerelease {
		args = append(args, "--prerelease")
	}
	if req.Accelerated {
		args = append(args, "--accelerated")
	}
	if req.Force {
		args = append(args, "--force")
	}
	if err := s.start(args); err != nil {
		return fmt.Errorf("start update stage: %w", err)
	}
	s.running = true
	return nil
}

func (s *UpdateService) Apply() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.Status()
	if s.running || status.State == "running" {
		return errors.New("update already running")
	}
	if status.Phase != "staged" {
		return errors.New("no staged update available")
	}
	// systemd-run places the helper outside the gpanel service cgroup. The
	// helper can therefore stop gpanel without being reaped with the server.
	cmd := exec.Command("systemd-run", "--unit=gpanel-online-apply-"+fmt.Sprint(time.Now().UnixNano()), "--collect", "--no-block", gpctlPath(), "online-apply")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("start update helper: %w", err)
	}
	s.running = true
	return nil
}

func (s *UpdateService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
