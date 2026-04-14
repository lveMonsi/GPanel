package service

import (
	"bufio"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"os"
	"path/filepath"
	"strings"
)

type SystemLogService struct{}

func NewSystemLogService() *SystemLogService {
	return &SystemLogService{}
}

// Search 读取日志文件，支持关键词过滤和倒序分页
func (s *SystemLogService) Search(req dto.SystemLogSearch) ([]string, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 100
	}

	logPath := s.resolveLogFile(req.LogFile)

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, 0, nil
		}
		return nil, 0, err
	}
	defer file.Close()

	// 读取所有行（带过滤）
	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if req.Keyword != "" && !strings.Contains(line, req.Keyword) {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	total := int64(len(lines))

	// 倒序：最新的在前面
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	// 分页
	start := (req.Page - 1) * req.PageSize
	if start >= len(lines) {
		return []string{}, total, nil
	}
	end := start + req.PageSize
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end], total, nil
}

// GetInfo 获取日志文件信息
func (s *SystemLogService) GetInfo() (*dto.SystemLogInfo, error) {
	info := &dto.SystemLogInfo{}

	logFiles := map[string]string{
		"agent.log":  global.AgentLogFile,
		"gpanel.log": filepath.Join(global.LogDirPath, "gpanel.log"),
	}

	for name, path := range logFiles {
		stat, err := os.Stat(path)
		if err != nil {
			continue
		}
		info.Files = append(info.Files, dto.SystemLogFile{
			Name: name,
			Path: path,
			Size: stat.Size(),
		})
	}

	return info, nil
}

func (s *SystemLogService) resolveLogFile(name string) string {
	switch name {
	case "gpanel.log":
		return filepath.Join(global.LogDirPath, "gpanel.log")
	default:
		return global.AgentLogFile
	}
}
