package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gpanel/agent/dto"
	"gpanel/agent/utils"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) GetFileList(req dto.FileListReq) (*dto.FileListRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
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
	op := utils.FileOption{Path: req.Path, Search: req.Search, ShowHidden: req.ShowHidden, Page: req.Page, PageSize: req.PageSize, SortBy: req.SortBy, SortOrder: req.SortOrder}
	return utils.NewFileList(op)
}

func (s *FileService) CreateFile(req dto.FileCreateReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
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
		mode = utils.GetFileMode(filepath.Dir(req.Path))
	}
	if req.IsDir {
		return fileOp.CreateDir(req.Path, mode|os.ModeDir)
	}
	return fileOp.CreateFile(req.Path, mode)
}

func (s *FileService) DeleteFile(req dto.FileDeleteReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
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

func (s *FileService) RenameFile(req dto.FileRenameReq) error {
	if req.OldPath == "" || req.NewPath == "" {
		return fmt.Errorf("oldPath and newPath are required")
	}
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

func (s *FileService) MoveFile(req dto.FileMoveReq) error {
	if len(req.OldPaths) == 0 || req.NewPath == "" {
		return fmt.Errorf("oldPaths and newPath are required")
	}
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
			} else if err := fileOp.CopyFile(oldPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %v", oldPath, err)
			}
		}
	}
	return nil
}

func (s *FileService) GetFileContent(req dto.FileContentReq) (*dto.FileContentRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
		}
		req.Path = absPath
	}
	return utils.GetFileContent(req.Path)
}

func (s *FileService) SaveFileContent(req dto.FileEditReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
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
	return fileOp.WriteFile(req.Path, req.Content, utils.GetFileMode(req.Path))
}

func (s *FileService) GetDirSize(req dto.DirSizeReq) (*dto.DirSizeRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
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
	return &dto.DirSizeRes{Path: req.Path, Size: size}, nil
}

func (s *FileService) CompressFiles(req dto.FileCompressReq) (string, error) {
	if len(req.Files) == 0 || req.Dst == "" || req.Name == "" {
		return "", fmt.Errorf("files, dst and name are required")
	}
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
	progressKey := utils.GenerateProgressKey("compress")
	pm := utils.GetProgressManager()
	pm.CreateProgress(progressKey, "压缩文件", 100)
	pm.SetProgressStatus(progressKey, "running")
	go func() {
		fileOp := utils.NewFileOp()
		cType := utils.CompressType(req.Type)
		if cType == "" {
			cType = utils.TarGz
		}
		pm.SetProgressMessage(progressKey, "正在压缩...")
		if err := fileOp.CompressFiles(req.Files, req.Dst, req.Name, cType); err != nil {
			pm.SetProgressStatus(progressKey, "failed")
			pm.SetProgressMessage(progressKey, err.Error())
			return
		}
		pm.UpdateProgress(progressKey, 100)
		pm.SetProgressStatus(progressKey, "completed")
		pm.SetProgressMessage(progressKey, "压缩完成")
	}()
	return progressKey, nil
}

func (s *FileService) DecompressFile(req dto.FileDecompressReq) (string, error) {
	if req.Path == "" || req.Dst == "" {
		return "", fmt.Errorf("path and dst are required")
	}
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
	progressKey := utils.GenerateProgressKey("decompress")
	pm := utils.GetProgressManager()
	pm.CreateProgress(progressKey, "解压文件", 100)
	pm.SetProgressStatus(progressKey, "running")
	go func() {
		pm.SetProgressMessage(progressKey, "正在解压...")
		if err := fileOp.DecompressFile(req.Path, req.Dst, cType); err != nil {
			pm.SetProgressStatus(progressKey, "failed")
			pm.SetProgressMessage(progressKey, err.Error())
			return
		}
		pm.UpdateProgress(progressKey, 100)
		pm.SetProgressStatus(progressKey, "completed")
		pm.SetProgressMessage(progressKey, "解压完成")
	}()
	return progressKey, nil
}

func (s *FileService) ChmodFile(req dto.FileChmodReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
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

func (s *FileService) ChownFile(req dto.FileChownReq) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
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

func (s *FileService) GetProgress(key string) *dto.ProgressInfo {
	return utils.GetProgressManager().GetProgress(key)
}

func (s *FileService) PreviewFile(req dto.FilePreviewReq) (*dto.FilePreviewRes, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
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
	result := &dto.FilePreviewRes{Path: req.Path, Name: info.Name(), MimeType: utils.GetMimeType(req.Path), Size: info.Size()}
	ext := strings.ToLower(filepath.Ext(req.Path))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".bmp"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg"}
	if utils.Contains(imageExts, ext) {
		result.Type = "image"
		content, err := fileOp.ReadFile(req.Path)
		if err != nil {
			return nil, err
		}
		result.Content = utils.Base64Encode(content)
	} else if utils.Contains(videoExts, ext) {
		result.Type = "video"
	} else if utils.Contains(audioExts, ext) {
		result.Type = "audio"
	} else {
		content, err := fileOp.ReadFile(req.Path)
		if err != nil {
			return nil, err
		}
		if !utils.IsTextExtension(req.Path) && len(content) > 0 && utils.DetectBinary(content) {
			result.Type = "binary"
		} else {
			result.Type = "text"
			result.Content = string(content)
		}
	}
	return result, nil
}

func (s *FileService) RemoteDownload(req dto.RemoteDownloadReq) (*dto.RemoteDownloadRes, error) {
	if req.URL == "" || req.Path == "" || req.Name == "" {
		return nil, fmt.Errorf("url, path and name are required")
	}
	if !filepath.IsAbs(req.Path) {
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert path to absolute path: %v", err)
		}
		req.Path = absPath
	}
	fileOp := utils.NewFileOp()
	if !fileOp.Stat(req.Path) {
		if err := fileOp.CreateDir(req.Path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}
	dstPath := filepath.Join(req.Path, req.Name)
	progressKey := utils.GenerateProgressKey("download")
	pm := utils.GetProgressManager()
	pm.CreateProgress(progressKey, "远程下载", 100)
	pm.SetProgressStatus(progressKey, "running")
	pm.SetProgressMessage(progressKey, "正在下载...")
	go func() {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			args := []string{"-L", "-o", dstPath, req.URL}
			if req.InsecureSkipVerify {
				args = append([]string{"-k"}, args...)
			}
			cmd = exec.Command("curl", args...)
		} else {
			args := []string{"-O", dstPath}
			if req.InsecureSkipVerify {
				args = append(args, "--no-check-certificate")
			}
			args = append(args, req.URL)
			cmd = exec.Command("wget", args...)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			pm.SetProgressStatus(progressKey, "failed")
			pm.SetProgressMessage(progressKey, fmt.Sprintf("下载失败: %v, output: %s", err, string(output)))
			return
		}
		pm.UpdateProgress(progressKey, 100)
		pm.SetProgressStatus(progressKey, "completed")
		pm.SetProgressMessage(progressKey, "下载完成")
	}()
	return &dto.RemoteDownloadRes{Key: progressKey, Path: dstPath}, nil
}
