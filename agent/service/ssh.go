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
	"sort"
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

func systemctlShowProperties(unit string, properties ...string) map[string]string {
	args := []string{"show", unit}
	for _, property := range properties {
		args = append(args, "--property="+property)
	}

	output, err := exec.Command("systemctl", args...).Output()
	if err != nil {
		return nil
	}

	values := make(map[string]string, len(properties))
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return values
}

func sshServiceUnit() string {
	for _, candidate := range []string{"sshd.service", "ssh.service", "sshd", "ssh"} {
		properties := systemctlShowProperties(candidate, "LoadState", "Id")
		if len(properties) == 0 {
			continue
		}
		if properties["LoadState"] == "not-found" {
			continue
		}

		serviceID := properties["Id"]
		if serviceID != "" {
			if strings.HasSuffix(serviceID, ".service") {
				return serviceID
			}
			return serviceID + ".service"
		}
	}
	return "sshd.service"
}

func sshAutoStartEnabled(svc string) bool {
	output, _ := exec.Command("systemctl", "status", svc, "--no-pager").CombinedOutput()
	statusOutput := strings.ToLower(string(output))
	for _, state := range []string{"enabled", "linked", "enabled-runtime"} {
		if strings.Contains(statusOutput, "; "+state+";") {
			return true
		}
	}

	output, _ = exec.Command("systemctl", "is-enabled", svc).CombinedOutput()
	enabledStatus := strings.TrimSpace(strings.ToLower(string(output)))
	return enabledStatus == "enabled" || enabledStatus == "linked" || enabledStatus == "enabled-runtime"
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

	svc := sshServiceUnit()
	out, _ := exec.Command("systemctl", "is-active", svc).Output()
	info.IsActive = strings.TrimSpace(string(out)) == "active"

	info.AutoStart = sshAutoStartEnabled(svc)

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
	svc := sshServiceUnit()
	switch operation {
	case "start", "stop", "restart":
		output, err := exec.Command("systemctl", operation, svc).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl %s %s failed: %s, %s", operation, svc, err, strings.TrimSpace(string(output)))
		}
		return nil
	case "enable", "disable":
		output, err := exec.Command("systemctl", operation, svc).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl %s %s failed: %s, %s", operation, svc, err, strings.TrimSpace(string(output)))
		}
		return nil
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
	return nil
}

func (s *SSHService) GetSSHSessions(loginUser string) ([]dto.SSHSession, error) {
	whoSessions, err := collectWhoSSHSessions(loginUser)
	if err != nil {
		return nil, err
	}

	hostByPID := collectSSHSocketHosts()
	psSessions, err := collectSSHProcessSessions(loginUser, hostByPID)
	if err != nil {
		return nil, err
	}

	sessionsByPID := make(map[int]dto.SSHSession, len(psSessions)+len(whoSessions))
	for _, session := range psSessions {
		session.Source = "system"
		if session.Terminal != "" {
			if whoSession, ok := whoSessions[session.Terminal]; ok {
				if session.LoginTime == "" {
					session.LoginTime = whoSession.LoginTime
				}
				if session.Host == "" {
					session.Host = whoSession.Host
				}
			}
		}
		sessionsByPID[session.PID] = session
	}

	for _, session := range whoSessions {
		session.Source = "system"
		matched := false
		for _, existing := range sessionsByPID {
			if session.Terminal != "" && session.Terminal == existing.Terminal {
				matched = true
				break
			}
		}
		if !matched {
			sessionsByPID[session.PID] = session
		}
	}

	sessions := make([]dto.SSHSession, 0, len(sessionsByPID))
	for _, session := range sessionsByPID {
		sessions = append(sessions, session)
	}

	// 合并 WebSocket SSH 会话
	for _, ws := range global.GetWsSSHSessions() {
		if loginUser != "" && !strings.Contains(ws.Username, loginUser) {
			continue
		}
		sessions = append(sessions, dto.SSHSession{
			PID:       int(ws.ID),
			Username:  ws.Username,
			Terminal:  "websocket",
			Host:      fmt.Sprintf("%s:%d", ws.Host, ws.Port),
			LoginTime: ws.LoginTime,
			Source:    "websocket",
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].PID > sessions[j].PID
	})
	return sessions, nil
}

func collectWhoSSHSessions(loginUser string) (map[string]dto.SSHSession, error) {
	out, err := exec.Command("who", "-u").Output()
	if err != nil {
		return nil, err
	}

	sessions := make(map[string]dto.SSHSession)
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+\s+\S+)\s+\S+\s+(\d+)\s+\(([^)]+)\)`)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		if loginUser != "" && !strings.Contains(m[1], loginUser) {
			continue
		}
		pid, _ := strconv.Atoi(m[4])
		sessions[m[2]] = dto.SSHSession{
			PID:       pid,
			Username:  m[1],
			Terminal:  m[2],
			LoginTime: m[3],
			Host:      m[5],
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func collectSSHProcessSessions(loginUser string, hostByPID map[int]string) ([]dto.SSHSession, error) {
	out, err := exec.Command("ps", "-eo", "pid=,user=,tty=,args=").Output()
	if err != nil {
		return nil, err
	}

	var sessions []dto.SSHSession
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		username, terminal, ok := parseSSHDSessionCommand(strings.Join(fields[3:], " "))
		if !ok {
			continue
		}
		if loginUser != "" && !strings.Contains(username, loginUser) {
			continue
		}
		if terminal == "?" {
			terminal = ""
		}
		sessions = append(sessions, dto.SSHSession{
			PID:       pid,
			Username:  username,
			Terminal:  terminal,
			Host:      hostByPID[pid],
			LoginTime: sshProcessStartTime(pid),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// collectSSHSocketHosts returns the peer address visible to sshd.
// When an SSH connection is relayed by a plain TCP proxy such as frpc, this
// is the proxy's address because the original source address is not present
// on the connection to sshd. Do not substitute an untrusted forwarded header.
func collectSSHSocketHosts() map[int]string {
	out, err := exec.Command("ss", "-tnp").Output()
	if err != nil {
		return map[int]string{}
	}

	hostByPID := make(map[int]string)
	pidRe := regexp.MustCompile(`pid=(\d+)`)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "ESTAB") || !strings.Contains(line, "sshd") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		peerHost := parseSocketHost(fields[4])
		if peerHost == "" {
			continue
		}
		for _, match := range pidRe.FindAllStringSubmatch(line, -1) {
			pid, err := strconv.Atoi(match[1])
			if err == nil {
				hostByPID[pid] = peerHost
			}
		}
	}
	return hostByPID
}

func parseSSHDSessionCommand(command string) (string, string, bool) {
	if !strings.HasPrefix(command, "sshd: ") {
		return "", "", false
	}
	session := strings.TrimSpace(strings.TrimPrefix(command, "sshd: "))
	if session == "" || strings.Contains(session, "listener") || strings.Contains(session, "accepting connections") {
		return "", "", false
	}
	if strings.Contains(session, "[priv]") || strings.Contains(session, "[net]") {
		return "", "", false
	}
	parts := strings.SplitN(session, "@", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	username := strings.TrimSpace(parts[0])
	terminal := strings.Fields(parts[1])
	if username == "" || len(terminal) == 0 {
		return "", "", false
	}
	return username, terminal[0], true
}

func sshProcessStartTime(pid int) string {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func parseSocketHost(address string) string {
	address = strings.TrimSpace(address)
	if address == "" || address == "*" {
		return ""
	}
	if strings.HasPrefix(address, "[") {
		if end := strings.Index(address, "]"); end > 1 {
			address = address[1:end]
		}
	} else if index := strings.LastIndex(address, ":"); index > 0 {
		address = address[:index]
	}
	address = strings.TrimPrefix(address, "::ffff:")
	return address
}

func (s *SSHService) KillSSHSession(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid")
	}
	// 先尝试关闭 WebSocket SSH 会话
	if global.CloseWsSSHSession(int64(pid)) {
		return nil
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
		out, err := exec.Command("journalctl", "-u", sshServiceUnit(), "--no-pager", "-n", "10000").Output()
		if err == nil {
			lines = strings.Split(string(out), "\n")
		}
	}

	acceptedRe := regexp.MustCompile(`Accepted (password|publickey|keyboard-interactive) for (\S+) from (\S+) port (\d+)`)
	failedRe := regexp.MustCompile(`Failed (password|publickey|keyboard-interactive) for (?:invalid user )?(\S+) from (\S+) port (\d+)`)
	invalidRe := regexp.MustCompile(`Invalid user (\S+) from (\S+) port (\d+)`)
	dateRe := regexp.MustCompile(`^(\w+\s+\d+\s+\d+:\d+:\d+)`)
	journalctlDateRe := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)

	currentYear := time.Now().Year()
	parseLogDate := func(line string) string {
		if m := journalctlDateRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
		if m := dateRe.FindStringSubmatch(line); m != nil {
			parsed, err := time.Parse("Jan 2 15:04:05 2006", m[1]+fmt.Sprintf(" %d", currentYear))
			if err == nil {
				return parsed.Format("2006-01-02 15:04:05")
			}
			return m[1]
		}
		return ""
	}

	var items []dto.SSHLogItem
	for _, line := range lines {
		var item dto.SSHLogItem
		date := parseLogDate(line)
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
