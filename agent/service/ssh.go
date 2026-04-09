package service

import (
	"bufio"
	"errors"
	"fmt"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/models"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const sshdConfig = "/etc/ssh/sshd_config"

type SSHService struct{}

func NewSSHService() *SSHService {
	return &SSHService{}
}

func sshServiceName() string {
	if err := exec.Command("systemctl", "is-active", "--quiet", "sshd").Run(); err == nil {
		return "sshd"
	}
	return "ssh"
}

func (s *SSHService) GetSSHInfo() (dto.SSHInfo, error) {
	info := dto.SSHInfo{
		Port:            22,
		ListenAddress:   "0.0.0.0",
		PasswordAuth:    "yes",
		PubkeyAuth:      "yes",
		PermitRootLogin: "yes",
		UseDNS:          "yes",
	}

	svc := sshServiceName()
	out, _ := exec.Command("systemctl", "is-active", svc).Output()
	info.IsActive = strings.TrimSpace(string(out)) == "active"

	out, _ = exec.Command("systemctl", "is-enabled", svc).Output()
	info.AutoStart = strings.TrimSpace(string(out)) == "enabled"

	f, err := os.Open(sshdConfig)
	if err != nil {
		return info, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "Port":
			if p, err := strconv.Atoi(val); err == nil {
				info.Port = p
			}
		case "ListenAddress":
			info.ListenAddress = val
		case "PasswordAuthentication":
			info.PasswordAuth = strings.ToLower(val)
		case "PubkeyAuthentication":
			info.PubkeyAuth = strings.ToLower(val)
		case "PermitRootLogin":
			info.PermitRootLogin = strings.ToLower(val)
		case "UseDNS":
			info.UseDNS = strings.ToLower(val)
		}
	}
	return info, nil
}

func (s *SSHService) OperateSSH(operation string) error {
	svc := sshServiceName()
	switch operation {
	case "start", "stop", "restart", "enable", "disable":
		return exec.Command("systemctl", operation, svc).Run()
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}
}

func (s *SSHService) UpdateSSHConfig(key, value string) error {
	data, err := os.ReadFile(sshdConfig)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	re := regexp.MustCompile(`(?i)^#?\s*` + regexp.QuoteMeta(key) + `\s+`)
	for i, line := range lines {
		if re.MatchString(line) {
			lines[i] = key + " " + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+" "+value)
	}

	return os.WriteFile(sshdConfig, []byte(strings.Join(lines, "\n")), 0644)
}

func (s *SSHService) LoadSSHFile(name string) (string, error) {
	path, err := sshFilePath(name)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *SSHService) UpdateSSHFile(name, value string) error {
	path, err := sshFilePath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	perm := os.FileMode(0644)
	if name == "authKeys" {
		perm = 0600
	}
	return os.WriteFile(path, []byte(value), perm)
}

func (s *SSHService) SearchSSHKeys(page, pageSize int) (dto.SSHKeySearchRes, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	if err := global.DB.Model(&models.SSHKey{}).Count(&total).Error; err != nil {
		return dto.SSHKeySearchRes{}, err
	}

	var keys []models.SSHKey
	offset := (page - 1) * pageSize
	if err := global.DB.Order("id desc").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return dto.SSHKeySearchRes{}, err
	}

	items := make([]dto.SSHKeyInfo, 0, len(keys))
	for _, item := range keys {
		items = append(items, toSSHKeyInfo(item))
	}
	return dto.SSHKeySearchRes{Total: total, Items: items}, nil
}

func (s *SSHService) CreateSSHKey(req dto.SSHKeyOperateReq) error {
	key, err := s.buildSSHKeyModel(req, false)
	if err != nil {
		return err
	}
	return global.DB.Create(key).Error
}

func (s *SSHService) UpdateSSHKey(req dto.SSHKeyOperateReq) error {
	if req.ID == 0 {
		return fmt.Errorf("invalid key id")
	}
	var existing models.SSHKey
	if err := global.DB.First(&existing, req.ID).Error; err != nil {
		return err
	}
	key, err := s.buildSSHKeyModel(req, true)
	if err != nil {
		return err
	}
	existing.Name = key.Name
	existing.Mode = key.Mode
	existing.EncryptionMode = key.EncryptionMode
	existing.PassPhrase = key.PassPhrase
	existing.Description = key.Description
	existing.PublicKey = key.PublicKey
	existing.PrivateKey = key.PrivateKey
	return global.DB.Save(&existing).Error
}

func (s *SSHService) DeleteSSHKeys(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.Delete(&models.SSHKey{}, ids).Error
}

func (s *SSHService) SyncSSHKeys() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".pub") {
			continue
		}
		path := filepath.Join(sshDir, name)
		privateKey, err := os.ReadFile(path)
		if err != nil || !looksLikePrivateKey(privateKey) {
			continue
		}
		publicKeyPath := path + ".pub"
		publicKey, _ := os.ReadFile(publicKeyPath)
		seen[name] = struct{}{}

		var existing models.SSHKey
		err = global.DB.Where("name = ?", name).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				key := models.SSHKey{
					Name:           name,
					Mode:           "sync",
					EncryptionMode: detectEncryptionMode(string(privateKey), name),
					Description:    "synced from ~/.ssh",
					PrivateKey:     string(privateKey),
					PublicKey:      string(publicKey),
				}
				if createErr := global.DB.Create(&key).Error; createErr != nil {
					return createErr
				}
				continue
			}
			return err
		}
		existing.Mode = "sync"
		existing.EncryptionMode = detectEncryptionMode(string(privateKey), name)
		existing.PrivateKey = string(privateKey)
		existing.PublicKey = string(publicKey)
		if existing.Description == "" {
			existing.Description = "synced from ~/.ssh"
		}
		if saveErr := global.DB.Save(&existing).Error; saveErr != nil {
			return saveErr
		}
	}
	_ = seen
	return nil
}

func (s *SSHService) GetSSHSessions() ([]dto.SSHSession, error) {
	out, err := exec.Command("who", "-u").Output()
	if err != nil {
		return nil, err
	}

	var sessions []dto.SSHSession
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+\s+\S+)\s+\S+\s+(\d+)\s+\(([^)]+)\)`)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[4])
		sessions = append(sessions, dto.SSHSession{
			PID:       pid,
			Username:  m[1],
			Terminal:  m[2],
			LoginTime: m[3],
			Host:      m[5],
		})
	}
	return sessions, nil
}

func (s *SSHService) KillSSHSession(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid")
	}
	return exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
}

func (s *SSHService) GetSSHLogs(page, pageSize int, status, info string) (dto.SSHLogRes, error) {
	var lines []string

	for _, f := range []string{"/var/log/auth.log.1", "/var/log/auth.log", "/var/log/secure"} {
		data, err := os.ReadFile(f)
		if err == nil {
			lines = append(lines, strings.Split(string(data), "\n")...)
		}
	}

	if len(lines) == 0 {
		out, err := exec.Command("journalctl", "-u", "sshd", "--no-pager", "-n", "10000").Output()
		if err == nil {
			lines = strings.Split(string(out), "\n")
		}
	}

	acceptedRe := regexp.MustCompile(`Accepted (password|publickey|keyboard-interactive) for (\S+) from (\S+) port (\d+)`)
	failedRe := regexp.MustCompile(`Failed (password|publickey|keyboard-interactive) for (?:invalid user )?(\S+) from (\S+) port (\d+)`)
	invalidRe := regexp.MustCompile(`Invalid user (\S+) from (\S+) port (\d+)`)
	dateRe := regexp.MustCompile(`^(\w+\s+\d+\s+\d+:\d+:\d+)`)

	var items []dto.SSHLogItem
	for _, line := range lines {
		var item dto.SSHLogItem
		date := ""
		if m := dateRe.FindStringSubmatch(line); m != nil {
			date = m[1]
		}
		if m := acceptedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: m[1], User: m[2], Address: m[3], Port: m[4], Status: "success"}
		} else if m := failedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: m[1], User: m[2], Address: m[3], Port: m[4], Status: "failed"}
		} else if m := invalidRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: "password", User: m[1], Address: m[2], Port: m[3], Status: "failed"}
		} else {
			continue
		}

		if status != "" && status != "all" && item.Status != status {
			continue
		}
		if info != "" && !strings.Contains(item.Address, info) && !strings.Contains(item.User, info) {
			continue
		}
		items = append(items, item)
	}

	total := int64(len(items))
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return dto.SSHLogRes{Total: total, Items: items[start:end]}, nil
}

func sshFilePath(name string) (string, error) {
	switch name {
	case "sshdConf":
		return sshdConfig, nil
	case "authKeys":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".ssh", "authorized_keys"), nil
	default:
		return "", fmt.Errorf("unsupported ssh file: %s", name)
	}
}

func (s *SSHService) buildSSHKeyModel(req dto.SSHKeyOperateReq, isUpdate bool) (*models.SSHKey, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "input"
	}
	encryptionMode := strings.TrimSpace(req.EncryptionMode)
	if encryptionMode == "" {
		encryptionMode = "ed25519"
	}

	privateKey := req.PrivateKey
	publicKey := req.PublicKey
	if mode == "generate" {
		generatedPrivate, generatedPublic, err := generateSSHKeyPair(name, encryptionMode, req.PassPhrase)
		if err != nil {
			return nil, err
		}
		privateKey = generatedPrivate
		publicKey = generatedPublic
	}
	if strings.TrimSpace(privateKey) == "" {
		return nil, fmt.Errorf("privateKey is required")
	}
	if strings.TrimSpace(publicKey) == "" {
		publicKey = derivePublicKey(privateKey)
	}
	if strings.TrimSpace(publicKey) == "" {
		return nil, fmt.Errorf("publicKey is required")
	}

	key := &models.SSHKey{
		Name:           name,
		Mode:           mode,
		EncryptionMode: encryptionMode,
		PassPhrase:     req.PassPhrase,
		Description:    req.Description,
		PublicKey:      publicKey,
		PrivateKey:     privateKey,
	}
	if isUpdate {
		key.BaseModel.ID = req.ID
	}
	return key, nil
}

func toSSHKeyInfo(item models.SSHKey) dto.SSHKeyInfo {
	return dto.SSHKeyInfo{
		ID:             item.ID,
		CreatedAt:      item.CreatedAt.Format(time.RFC3339),
		Name:           item.Name,
		Mode:           item.Mode,
		EncryptionMode: item.EncryptionMode,
		PassPhrase:     item.PassPhrase,
		Description:    item.Description,
		PublicKey:      item.PublicKey,
		PrivateKey:     item.PrivateKey,
	}
}

func looksLikePrivateKey(content []byte) bool {
	text := string(content)
	return strings.Contains(text, "BEGIN OPENSSH PRIVATE KEY") || strings.Contains(text, "BEGIN RSA PRIVATE KEY") || strings.Contains(text, "BEGIN EC PRIVATE KEY") || strings.Contains(text, "BEGIN DSA PRIVATE KEY")
}

func detectEncryptionMode(privateKey string, name string) string {
	upper := strings.ToUpper(privateKey)
	switch {
	case strings.Contains(upper, "OPENSSH PRIVATE KEY") && strings.Contains(strings.ToLower(name), "ed25519"):
		return "ed25519"
	case strings.Contains(upper, "RSA PRIVATE KEY"):
		return "rsa"
	case strings.Contains(upper, "EC PRIVATE KEY"):
		return "ecdsa"
	case strings.Contains(upper, "DSA PRIVATE KEY"):
		return "dsa"
	case strings.Contains(strings.ToLower(name), "rsa"):
		return "rsa"
	case strings.Contains(strings.ToLower(name), "ecdsa"):
		return "ecdsa"
	case strings.Contains(strings.ToLower(name), "dsa"):
		return "dsa"
	default:
		return "ed25519"
	}
}

func derivePublicKey(privateKey string) string {
	trimmed := strings.TrimSpace(privateKey)
	if trimmed == "" {
		return ""
	}
	return ""
}

func generateSSHKeyPair(name, encryptionMode, passPhrase string) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "gpanel-ssh-key-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, sanitizeFileName(name))
	args := []string{"-q", "-f", keyPath, "-N", passPhrase}
	switch encryptionMode {
	case "rsa":
		args = append(args, "-t", "rsa")
	case "ecdsa":
		args = append(args, "-t", "ecdsa")
	case "dsa":
		args = append(args, "-t", "dsa")
	default:
		args = append(args, "-t", "ed25519")
	}
	if output, err := exec.Command("ssh-keygen", args...).CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("ssh-keygen failed: %s", strings.TrimSpace(string(output)))
	}
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return "", "", err
	}
	publicKey, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return "", "", err
	}
	return string(privateKey), string(publicKey), nil
}

func sanitizeFileName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	clean := re.ReplaceAllString(name, "_")
	if clean == "" {
		return "id_key"
	}
	return clean
}
