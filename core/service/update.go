package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
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

type ReleaseChannel string

const (
	StableChannel     ReleaseChannel = "stable"
	PrereleaseChannel ReleaseChannel = "prerelease"
	githubAPIBase                    = "https://api.github.com/repos/lveMonsi/GPanel"
)

var (
	ErrInvalidReleaseChannel = errors.New("invalid release channel")
	ErrNoRelease             = errors.New("no valid release found")
)

type LatestRelease struct {
	Version     string         `json:"version"`
	Channel     ReleaseChannel `json:"channel"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt string         `json:"publishedAt"`
	HTMLURL     string         `json:"htmlUrl,omitempty"`
}

func ClassifyReleaseChannel(version string) ReleaseChannel {
	if strings.HasPrefix(version, "pre-release-") {
		return PrereleaseChannel
	}
	if strings.HasPrefix(version, "v") {
		return StableChannel
	}
	return ""
}

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	Published  string `json:"published_at"`
	HTMLURL    string `json:"html_url"`
}

type ReleaseClient interface {
	Latest(context.Context, ReleaseChannel) (LatestRelease, error)
}

type githubReleaseClient struct {
	httpClient *http.Client
	apiBase    string
}

func newGitHubReleaseClient() ReleaseClient {
	return &githubReleaseClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiBase:    githubAPIBase,
	}
}

func (c *githubReleaseClient) Latest(ctx context.Context, channel ReleaseChannel) (LatestRelease, error) {
	if channel != StableChannel && channel != PrereleaseChannel {
		return LatestRelease{}, ErrInvalidReleaseChannel
	}

	endpoint := c.apiBase + "/releases/latest"
	if channel == PrereleaseChannel {
		endpoint = c.apiBase + "/releases?per_page=100"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return LatestRelease{}, fmt.Errorf("create release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gpanel-update")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return LatestRelease{}, fmt.Errorf("request releases: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20+1))
	if err != nil {
		return LatestRelease{}, fmt.Errorf("read release response: %w", err)
	}
	if len(body) > 4<<20 {
		return LatestRelease{}, errors.New("release response too large")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return LatestRelease{}, fmt.Errorf("github returned HTTP %s", resp.Status)
	}

	if channel == StableChannel {
		var release githubRelease
		if err := json.Unmarshal(body, &release); err != nil {
			return LatestRelease{}, fmt.Errorf("decode stable release: %w", err)
		}
		if !isReleaseForChannel(release, channel) {
			return LatestRelease{}, ErrNoRelease
		}
		return makeLatestRelease(release, channel), nil
	}

	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return LatestRelease{}, fmt.Errorf("decode prerelease list: %w", err)
	}
	valid := make([]githubRelease, 0, len(releases))
	for _, release := range releases {
		if isReleaseForChannel(release, channel) {
			if _, err := time.Parse(time.RFC3339, release.Published); err != nil {
				continue
			}
			valid = append(valid, release)
		}
	}
	if len(valid) == 0 {
		return LatestRelease{}, ErrNoRelease
	}
	sort.SliceStable(valid, func(i, j int) bool {
		left, leftErr := time.Parse(time.RFC3339, valid[i].Published)
		right, rightErr := time.Parse(time.RFC3339, valid[j].Published)
		if leftErr != nil || rightErr != nil {
			return valid[i].Published > valid[j].Published
		}
		return left.After(right)
	})
	return makeLatestRelease(valid[0], channel), nil
}

func isReleaseForChannel(release githubRelease, channel ReleaseChannel) bool {
	if release.Draft {
		return false
	}
	if channel == StableChannel {
		return !release.Prerelease && strings.HasPrefix(release.TagName, "v")
	}
	return release.Prerelease && strings.HasPrefix(release.TagName, "pre-release-")
}

func makeLatestRelease(release githubRelease, channel ReleaseChannel) LatestRelease {
	return LatestRelease{
		Version: release.TagName, Channel: channel, Prerelease: release.Prerelease,
		PublishedAt: release.Published, HTMLURL: release.HTMLURL,
	}
}

type UpdateStatus struct {
	State         string         `json:"state"`
	Phase         string         `json:"phase"`
	Channel       ReleaseChannel `json:"channel,omitempty"`
	Version       string         `json:"version,omitempty"`
	TargetVersion string         `json:"targetVersion,omitempty"`
	Message       string         `json:"message,omitempty"`
	Percent       int            `json:"percent"`
	StartedAt     time.Time      `json:"startedAt,omitempty"`
	FinishedAt    time.Time      `json:"finishedAt,omitempty"`
	UpdatedAt     time.Time      `json:"updatedAt,omitempty"`
	Error         string         `json:"error,omitempty"`
}

type onlineStatus struct {
	Phase         string         `json:"phase"`
	Message       string         `json:"message"`
	Channel       ReleaseChannel `json:"channel,omitempty"`
	Version       string         `json:"version,omitempty"`
	TargetVersion string         `json:"target_version,omitempty"`
	Percent       int            `json:"percent"`
	Error         string         `json:"error,omitempty"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type UpdateService struct {
	mu            sync.Mutex
	running       bool
	releaseClient ReleaseClient
}

func NewUpdateService() *UpdateService {
	return NewUpdateServiceWithReleaseClient(newGitHubReleaseClient())
}

func NewUpdateServiceWithReleaseClient(client ReleaseClient) *UpdateService {
	return &UpdateService{releaseClient: client}
}

func (s *UpdateService) Latest(ctx context.Context, channel ReleaseChannel) (LatestRelease, error) {
	if channel != StableChannel && channel != PrereleaseChannel {
		return LatestRelease{}, ErrInvalidReleaseChannel
	}
	return s.releaseClient.Latest(ctx, channel)
}

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
	channel := raw.Channel
	if channel == "" {
		channel = ClassifyReleaseChannel(raw.TargetVersion)
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
		State: state, Phase: raw.Phase, Channel: channel, Version: raw.Version,
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
