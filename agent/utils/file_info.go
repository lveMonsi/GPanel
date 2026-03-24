package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"gpanel/agent/dto"
)

type FileOption struct {
	Path       string
	Search     string
	ShowHidden bool
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}

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
	} else if end > total {
		result = items[start:]
	} else {
		result = items[start:end]
	}
	return &dto.FileListRes{Path: op.Path, Name: filepath.Base(op.Path), Items: result, ItemTotal: total}, nil
}

func convertToFileInfo(basePath string, info os.FileInfo) dto.FileInfo {
	fullPath := filepath.Join(basePath, info.Name())
	mode := info.Mode()
	user, group, uid, gid := GetFileOwner(fullPath)
	fileInfo := dto.FileInfo{Path: fullPath, Name: info.Name(), Size: info.Size(), IsDir: info.IsDir(), IsSymlink: IsSymlink(mode), IsHidden: IsHidden(fullPath), Extension: filepath.Ext(info.Name()), Mode: fmt.Sprintf("%04o", mode.Perm()), MimeType: GetMimeType(fullPath), ModTime: info.ModTime(), User: user, Group: group, Uid: strconv.FormatUint(uint64(uid), 10), Gid: strconv.FormatUint(uint64(gid), 10)}
	if IsSymlink(mode) {
		linkPath := GetSymlink(fullPath)
		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(filepath.Dir(fullPath), linkPath)
		}
		fileInfo.LinkPath = linkPath
		if targetInfo, err := os.Stat(linkPath); err == nil {
			fileInfo.IsDir = targetInfo.IsDir()
			fileInfo.Extension = filepath.Ext(linkPath)
		}
	}
	return fileInfo
}

func sortFiles(files []os.FileInfo, sortBy, sortOrder string) {
	switch sortBy {
	case "name":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
		} else {
			sort.Slice(files, func(i, j int) bool { return files[i].Name() > files[j].Name() })
		}
	case "size":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool { return files[i].Size() < files[j].Size() })
		} else {
			sort.Slice(files, func(i, j int) bool { return files[i].Size() > files[j].Size() })
		}
	case "modTime":
		if sortOrder == "ascending" {
			sort.Slice(files, func(i, j int) bool { return files[i].ModTime().Before(files[j].ModTime()) })
		} else {
			sort.Slice(files, func(i, j int) bool { return files[i].ModTime().After(files[j].ModTime()) })
		}
	default:
		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	}
}

func GetFileContent(path string) (*dto.FileContentRes, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}
	const maxFileSize = 10 * 1024 * 1024
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var content []byte
	var truncated bool
	headBuf := make([]byte, 8192)
	n, err := file.Read(headBuf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	headBuf = headBuf[:n]
	if !IsTextExtension(path) && len(headBuf) > 0 && DetectBinary(headBuf) {
		return nil, fmt.Errorf("file is binary and cannot be displayed")
	}
	if info.Size() <= maxFileSize {
		content = make([]byte, info.Size())
		copy(content, headBuf)
		if _, err := file.ReadAt(content[n:], int64(n)); err != nil && err != io.EOF {
			return nil, err
		}
	} else {
		lines, err := tailFromEnd(path, 300)
		if err != nil {
			return nil, err
		}
		content = []byte(strings.Join(lines, "\n"))
		truncated = true
	}
	content = convertToUTF8(content)
	return &dto.FileContentRes{Path: path, Name: info.Name(), Content: string(content), Mode: fmt.Sprintf("%04o", info.Mode().Perm()), Size: info.Size(), MimeType: GetMimeType(path), Truncated: truncated}, nil
}

func convertToUTF8(content []byte) []byte {
	if len(content) == 0 {
		return content
	}
	_, encodingName, _ := charset.DetermineEncoding(content, "")
	if encodingName == "" || encodingName == "utf-8" {
		return content
	}
	decoder := getDecoderByName(encodingName)
	if decoder == nil {
		return content
	}
	reader := transform.NewReader(bytes.NewReader(content), decoder.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return content
	}
	return decoded
}

func getDecoderByName(name string) encoding.Encoding {
	switch strings.ToLower(name) {
	case "gbk", "gb2312":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1
	case "windows-1252":
		return charmap.Windows1252
	default:
		return nil
	}
}

func tailFromEnd(path string, lines int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	if size == 0 {
		return []string{}, nil
	}
	buf := make([]byte, 4096)
	var content []byte
	lineCount := 0
	pos := size
	for pos > 0 && lineCount < lines {
		readSize := int64(len(buf))
		if pos < readSize {
			readSize = pos
		}
		pos -= readSize
		if _, err := file.ReadAt(buf[:readSize], pos); err != nil {
			return nil, err
		}
		content = append(buf[:readSize], content...)
		for i := 0; i < len(buf[:readSize]); i++ {
			if buf[i] == '\n' {
				lineCount++
				if lineCount >= lines {
					break
				}
			}
		}
	}
	var result []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if len(result) > lines {
		result = result[len(result)-lines:]
	}
	return result, nil
}
