package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gpanel/dto"
	"gpanel/utils"
)

// FileService 文件服务
type FileService struct{}

// NewFileService 创建文件服务
func NewFileService() *FileService {
	return &FileService{}
}

// GetFileList 获取文件列表
func (s *FileService) GetFileList(req dto.FileListReq) (*dto.FileListRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 100
	}

	op := utils.FileOption{
		Path:       req.Path,
		Search:     req.Search,
		ShowHidden: req.ShowHidden,
		Page:       req.Page,
		PageSize:   req.PageSize,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	return utils.NewFileList(op)
}

// CreateFile 创建文件或目录
func (s *FileService) CreateFile(req dto.FileCreateReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if fileOp.Stat(req.Path) {
		return fmt.Errorf("file or directory already exists")
	}

	mode := os.FileMode(req.Mode)
	if mode == 0 {
		parentDir := filepath.Dir(req.Path)
		mode = utils.GetFileMode(parentDir)
	}

	if req.IsDir {
		return fileOp.CreateDir(req.Path, mode|os.ModeDir)
	}

	return fileOp.CreateFile(req.Path, mode)
}

// DeleteFile 删除文件或目录
func (s *FileService) DeleteFile(req dto.FileDeleteReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return fmt.Errorf("file or directory not found")
	}

	info, err := os.Stat(req.Path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fileOp.DeleteDir(req.Path)
	}

	return fileOp.DeleteFile(req.Path)
}

// RenameFile 重命名文件或目录
func (s *FileService) RenameFile(req dto.FileRenameReq) error {
	if req.OldPath == "" || req.NewPath == "" {
		return fmt.Errorf("oldPath and newPath are required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.OldPath) {
		absPath, err := filepath.Abs(req.OldPath)
		if err != nil {
			return fmt.Errorf("failed to convert oldPath to absolute path: %v", err)
		}
		req.OldPath = absPath
	}

	if !filepath.IsAbs(req.NewPath) {
		absPath, err := filepath.Abs(req.NewPath)
		if err != nil {
			return fmt.Errorf("failed to convert newPath to absolute path: %v", err)
		}
		req.NewPath = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.OldPath) {
		return fmt.Errorf("file not found")
	}

	if fileOp.Stat(req.NewPath) {
		return fmt.Errorf("target file already exists")
	}

	return fileOp.Rename(req.OldPath, req.NewPath)
}

// MoveFile 移动或复制文件
func (s *FileService) MoveFile(req dto.FileMoveReq) error {
	if len(req.OldPaths) == 0 || req.NewPath == "" {
		return fmt.Errorf("oldPaths and newPath are required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.NewPath) {
		absPath, err := filepath.Abs(req.NewPath)
		if err != nil {
			return fmt.Errorf("failed to convert newPath to absolute path: %v", err)
		}
		req.NewPath = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.NewPath) {
		return fmt.Errorf("target directory not found")
	}

	newPathInfo, err := os.Stat(req.NewPath)
	if err != nil {
		return err
	}

	if !newPathInfo.IsDir() {
		return fmt.Errorf("newPath must be a directory")
	}

	for i, oldPath := range req.OldPaths {
		// 将相对路径转换为绝对路径
		if !filepath.IsAbs(oldPath) {
			absPath, err := filepath.Abs(oldPath)
			if err != nil {
				return fmt.Errorf("failed to convert oldPath to absolute path: %v", err)
			}
			req.OldPaths[i] = absPath
			oldPath = absPath
		}

		if !fileOp.Stat(oldPath) {
			return fmt.Errorf("file not found: %s", oldPath)
		}

		oldPathInfo, err := os.Stat(oldPath)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(req.NewPath, filepath.Base(oldPath))

		if fileOp.Stat(dstPath) && !req.Cover {
			return fmt.Errorf("target already exists: %s", dstPath)
		}

		if req.Type == "cut" {
			if err := fileOp.Rename(oldPath, dstPath); err != nil {
				return fmt.Errorf("failed to move %s: %v", oldPath, err)
			}
		} else if req.Type == "copy" {
			if oldPathInfo.IsDir() {
				if err := fileOp.CopyDir(oldPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy directory %s: %v", oldPath, err)
				}
			} else {
				if err := fileOp.CopyFile(oldPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy file %s: %v", oldPath, err)
				}
			}
		}
	}

	return nil
}

// GetFileContent 获取文件内容
func (s *FileService) GetFileContent(req dto.FileContentReq) (*dto.FileContentRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	return utils.GetFileContent(req.Path)
}

// SaveFileContent 保存文件内容
func (s *FileService) SaveFileContent(req dto.FileEditReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return fmt.Errorf("file not found")
	}

	mode := utils.GetFileMode(req.Path)
	return fileOp.WriteFile(req.Path, req.Content, mode)
}

// GetDirSize 获取目录大小
func (s *FileService) GetDirSize(req dto.DirSizeReq) (*dto.DirSizeRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	size, err := fileOp.GetDirSize(req.Path)
	if err != nil {
		return nil, err
	}

	return &dto.DirSizeRes{
		Path: req.Path,
		Size: size,
	}, nil
}

// ValidatePath 验证路径
func (s *FileService) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}

	if strings.Contains(path, "..") {
		cleaned := filepath.Clean(path)
		if strings.Contains(cleaned, "..") {
			return fmt.Errorf("invalid path")
		}
	}

	return nil
}

// GetAbsolutePath 获取绝对路径
func (s *FileService) GetAbsolutePath(basePath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(basePath, relativePath)
}

// CompressFiles 压缩文件
func (s *FileService) CompressFiles(req dto.FileCompressReq) (string, error) {
	if len(req.Files) == 0 {
		return "", fmt.Errorf("files is required")
	}
	if req.Dst == "" {
		return "", fmt.Errorf("dst is required")
	}
	if req.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Dst) {
		absPath, err := filepath.Abs(req.Dst)
		if err != nil {
			return "", fmt.Errorf("failed to convert dst to absolute path: %v", err)
		}
		req.Dst = absPath
	}

	for i, file := range req.Files {
		if !filepath.IsAbs(file) {
			absPath, err := filepath.Abs(file)
			if err != nil {
				return "", fmt.Errorf("failed to convert file path to absolute path: %v", err)
			}
			req.Files[i] = absPath
		}
	}

	// 创建进度跟踪
	progressKey := utils.GenerateProgressKey("compress")
	progressManager := utils.GetProgressManager()
	progressManager.CreateProgress(progressKey, "压缩文件", 100)
	progressManager.SetProgressStatus(progressKey, "running")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				progressManager.SetProgressStatus(progressKey, "failed")
				progressManager.SetProgressMessage(progressKey, fmt.Sprintf("压缩失败: %v", r))
			}
		}()

		fileOp := utils.NewFileOp()
		cType := utils.CompressType(req.Type)
		if cType == "" {
			cType = utils.TarGz
		}

		progressManager.SetProgressMessage(progressKey, "正在压缩...")
		if err := fileOp.CompressFiles(req.Files, req.Dst, req.Name, cType); err != nil {
			progressManager.SetProgressStatus(progressKey, "failed")
			progressManager.SetProgressMessage(progressKey, err.Error())
			return
		}

		progressManager.UpdateProgress(progressKey, 100)
		progressManager.SetProgressStatus(progressKey, "completed")
		progressManager.SetProgressMessage(progressKey, "压缩完成")
	}()

	return progressKey, nil
}

// DecompressFile 解压文件
func (s *FileService) DecompressFile(req dto.FileDecompressReq) (string, error) {
	if req.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if req.Dst == "" {
		return "", fmt.Errorf("dst is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return "", fmt.Errorf("failed to convert path to absolute path: %v", err)
		}
		req.Path = absPath
	}

	if !filepath.IsAbs(req.Dst) {
		absPath, err := filepath.Abs(req.Dst)
		if err != nil {
			return "", fmt.Errorf("failed to convert dst to absolute path: %v", err)
		}
		req.Dst = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return "", fmt.Errorf("source file not found")
	}

	cType := utils.CompressType(req.Type)
	if cType == "" {
		cType = utils.DetectCompressType(req.Path)
	}

	if cType == "" {
		return "", fmt.Errorf("unknown compress type")
	}

	// 创建进度跟踪
	progressKey := utils.GenerateProgressKey("decompress")
	progressManager := utils.GetProgressManager()
	progressManager.CreateProgress(progressKey, "解压文件", 100)
	progressManager.SetProgressStatus(progressKey, "running")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				progressManager.SetProgressStatus(progressKey, "failed")
				progressManager.SetProgressMessage(progressKey, fmt.Sprintf("解压失败: %v", r))
			}
		}()

		progressManager.SetProgressMessage(progressKey, "正在解压...")
		if err := fileOp.DecompressFile(req.Path, req.Dst, cType); err != nil {
			progressManager.SetProgressStatus(progressKey, "failed")
			progressManager.SetProgressMessage(progressKey, err.Error())
			return
		}

		progressManager.UpdateProgress(progressKey, 100)
		progressManager.SetProgressStatus(progressKey, "completed")
		progressManager.SetProgressMessage(progressKey, "解压完成")
	}()

	return progressKey, nil
}

// ChmodFile 修改文件权限
func (s *FileService) ChmodFile(req dto.FileChmodReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return fmt.Errorf("file not found")
	}

	return fileOp.ChmodR(req.Path, os.FileMode(req.Mode), req.Sub)
}

// ChownFile 修改文件所有者
func (s *FileService) ChownFile(req dto.FileChownReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return fmt.Errorf("file not found")
	}

	return fileOp.ChownR(req.Path, req.User, req.Group, req.Sub)
}

// GetProgress 获取进度信息
func (s *FileService) GetProgress(key string) *dto.ProgressInfo {
	progressManager := utils.GetProgressManager()
	return progressManager.GetProgress(key)
}

// PreviewFile 预览文件
func (s *FileService) PreviewFile(req dto.FilePreviewReq) (*dto.FilePreviewRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		return nil, fmt.Errorf("file not found")
	}

	info, err := os.Stat(req.Path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("cannot preview directory")
	}

	result := &dto.FilePreviewRes{
		Path:     req.Path,
		Name:     info.Name(),
		MimeType: utils.GetMimeType(req.Path),
		Size:     info.Size(),
	}

	// 检测文件类型
	ext := strings.ToLower(filepath.Ext(req.Path))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".bmp"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg"}

	if utils.Contains(imageExts, ext) {
		result.Type = "image"
		// 读取图片文件并转换为 base64
		content, err := fileOp.ReadFile(req.Path)
		if err != nil {
			return nil, err
		}
		result.Content = utils.Base64Encode(content)
	} else if utils.Contains(videoExts, ext) {
		result.Type = "video"
		result.Content = "" // 视频文件不返回内容，由前端直接播放
	} else if utils.Contains(audioExts, ext) {
		result.Type = "audio"
		result.Content = "" // 音频文件不返回内容，由前端直接播放
	} else {
		// 尝试作为文本文件处理
		content, err := fileOp.ReadFile(req.Path)
		if err != nil {
			return nil, err
		}

		if len(content) > 0 && utils.DetectBinary(content) {
			result.Type = "binary"
			result.Content = ""
		} else {
			result.Type = "text"
			result.Content = string(content)
		}
	}

	return result, nil
}