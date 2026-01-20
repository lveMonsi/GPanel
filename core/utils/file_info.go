package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gpanel/dto"
)

// FileOption 文件选项
type FileOption struct {
	Path       string
	Search     string
	ShowHidden bool
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}

// NewFileList 获取文件列表
func NewFileList(op FileOption) (*dto.FileListRes, error) {
	info, err := os.Stat(op.Path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		op.Path = filepath.Dir(op.Path)
	}

	entries, err := os.ReadDir(op.Path)
	if err != nil {
		return nil, err
	}

	var dirs []os.FileInfo
	var files []os.FileInfo

	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		if !op.ShowHidden && strings.HasPrefix(fileInfo.Name(), ".") {
			continue
		}

		if op.Search != "" && !strings.Contains(strings.ToLower(fileInfo.Name()), strings.ToLower(op.Search)) {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, fileInfo)
		} else {
			files = append(files, fileInfo)
		}
	}

	sortFiles(dirs, op.SortBy, op.SortOrder)
	sortFiles(files, op.SortBy, op.SortOrder)

	var items []dto.FileInfo
	for _, dir := range dirs {
		items = append(items, convertToFileInfo(op.Path, dir))
	}
	for _, file := range files {
		items = append(items, convertToFileInfo(op.Path, file))
	}

	total := len(items)
	start := (op.Page - 1) * op.PageSize
	end := op.PageSize + start

	var result []dto.FileInfo
	if start < 0 || start > total || end < 0 || start > end {
		result = items
	} else {
		if end > total {
			result = items[start:]
		} else {
			result = items[start:end]
		}
	}

	return &dto.FileListRes{
		Path:      op.Path,
		Name:      filepath.Base(op.Path),
		Items:     result,
		ItemTotal: total,
	}, nil
}

// convertToFileInfo 转换为文件信息
func convertToFileInfo(basePath string, info os.FileInfo) dto.FileInfo {
	fullPath := filepath.Join(basePath, info.Name())
	mode := info.Mode()
	user, group, uid, gid := GetFileOwner(fullPath)

	fileInfo := dto.FileInfo{
		Path:      fullPath,
		Name:      info.Name(),
		Size:      info.Size(),
		IsDir:     info.IsDir(),
		IsSymlink: IsSymlink(mode),
		IsHidden:  IsHidden(fullPath),
		Extension: filepath.Ext(info.Name()),
		Mode:      fmt.Sprintf("%04o", mode.Perm()),
		MimeType:  GetMimeType(fullPath),
		ModTime:   info.ModTime(),
		User:      user,
		Group:     group,
		Uid:       strconv.FormatUint(uint64(uid), 10),
		Gid:       strconv.FormatUint(uint64(gid), 10),
	}

	if IsSymlink(mode) {
		linkPath := GetSymlink(fullPath)
		if !filepath.IsAbs(linkPath) {
			dir := filepath.Dir(fullPath)
			linkPath = filepath.Join(dir, linkPath)
		}
		fileInfo.LinkPath = linkPath
		
		targetInfo, err := os.Stat(linkPath)
		if err == nil {
			fileInfo.IsDir = targetInfo.IsDir()
			fileInfo.Extension = filepath.Ext(linkPath)
		}
	}

	return fileInfo
}

// sortFiles 排序文件列表
func sortFiles(files []os.FileInfo, sortBy, sortOrder string) {
	switch sortBy {
	case "name":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() < files[j].Name()
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() > files[j].Name()
			})
		}
	case "size":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size() < files[j].Size()
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size() > files[j].Size()
			})
		}
	case "modTime":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].ModTime().Before(files[j].ModTime())
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].ModTime().After(files[j].ModTime())
			})
		}
	default:
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})
	}
}

// GetFileContent 获取文件内容
func GetFileContent(path string) (*dto.FileContentRes, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}

	fileOp := NewFileOp()
	content, err := fileOp.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(content) > 0 && DetectBinary(content) {
		return nil, fmt.Errorf("file is binary and cannot be displayed")
	}

	return &dto.FileContentRes{
		Path:     path,
		Name:     info.Name(),
		Content:  string(content),
		Mode:     fmt.Sprintf("%04o", info.Mode().Perm()),
		Size:     info.Size(),
		MimeType: GetMimeType(path),
	}, nil
}