package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gpanel/agent/dto"
	"gpanel/agent/service"
	"gpanel/agent/utils"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService *service.FileService
}

func NewFileController() *FileController {
	return &FileController{fileService: service.NewFileService()}
}

func (c *FileController) success(ctx *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "success"
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": message, "data": data})
}

func (c *FileController) fail(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

func (c *FileController) GetDrives(ctx *gin.Context) {
	var drives []string
	if runtime.GOOS == "windows" {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drivePath := string(drive) + ":\\"
			if _, err := os.Stat(drivePath); err == nil {
				drives = append(drives, drivePath)
			}
		}
	} else {
		drives = append(drives, "/")
	}
	c.success(ctx, drives, "success")
}

func (c *FileController) ListFiles(ctx *gin.Context) {
	var req dto.FileListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := c.fileService.GetFileList(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result, "success")
}

func (c *FileController) CreateFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileCreateReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.CreateFile(req)
	})
}
func (c *FileController) DeleteFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileDeleteReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.DeleteFile(req)
	})
}
func (c *FileController) RenameFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileRenameReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.RenameFile(req)
	})
}
func (c *FileController) MoveFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileMoveReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.MoveFile(req)
	})
}
func (c *FileController) SaveFileContent(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileEditReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.SaveFileContent(req)
	})
}
func (c *FileController) ChmodFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileChmodReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.ChmodFile(req)
	})
}
func (c *FileController) ChownFile(ctx *gin.Context) {
	c.handleJSON(ctx, func() error {
		var req dto.FileChownReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return err
		}
		return c.fileService.ChownFile(req)
	})
}

func (c *FileController) GetFileContent(ctx *gin.Context) {
	var req dto.FileContentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := c.fileService.GetFileContent(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result, "success")
}

func (c *FileController) GetDirSize(ctx *gin.Context) {
	var req dto.DirSizeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := c.fileService.GetDirSize(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result, "success")
}

func (c *FileController) CompressFiles(ctx *gin.Context) {
	var req dto.FileCompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	key, err := c.fileService.CompressFiles(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, gin.H{"key": key}, "success")
}

func (c *FileController) DecompressFile(ctx *gin.Context) {
	var req dto.FileDecompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	key, err := c.fileService.DecompressFile(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, gin.H{"key": key}, "success")
}

func (c *FileController) GetProgress(ctx *gin.Context) {
	key := ctx.Query("key")
	if key == "" {
		c.fail(ctx, http.StatusBadRequest, "key is required")
		return
	}
	progress := c.fileService.GetProgress(key)
	if progress == nil {
		c.fail(ctx, http.StatusNotFound, "progress not found")
		return
	}
	c.success(ctx, progress, "success")
}

func (c *FileController) PreviewFile(ctx *gin.Context) {
	var req dto.FilePreviewReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := c.fileService.PreviewFile(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result, "success")
}

func (c *FileController) RemoteDownload(ctx *gin.Context) {
	var req dto.RemoteDownloadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := c.fileService.RemoteDownload(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result, "success")
}

func (c *FileController) UploadFile(ctx *gin.Context) {
	path := ctx.PostForm("path")
	if path == "" {
		c.fail(ctx, http.StatusBadRequest, "path is required")
		return
	}
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			c.fail(ctx, http.StatusInternalServerError, "failed to convert to absolute path: "+err.Error())
			return
		}
		path = absPath
	}
	overwriteFlag := strings.ToLower(ctx.PostForm("overwrite")) == "true"
	fileOp := utils.NewFileOp()
	if !fileOp.Stat(path) {
		if err := fileOp.CreateDir(path, 0755); err != nil {
			c.fail(ctx, http.StatusInternalServerError, "failed to create directory: "+err.Error())
			return
		}
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		c.fail(ctx, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		c.fail(ctx, http.StatusBadRequest, "no files uploaded")
		return
	}
	successCount := 0
	for _, fileHeader := range files {
		dstPath := filepath.Join(path, fileHeader.Filename)
		if fileOp.Stat(dstPath) && !overwriteFlag {
			continue
		}
		srcFile, err := fileHeader.Open()
		if err != nil {
			continue
		}
		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			srcFile.Close()
			continue
		}
		_, copyErr := dstFile.ReadFrom(srcFile)
		dstFile.Close()
		srcFile.Close()
		if copyErr != nil {
			continue
		}
		successCount++
	}
	c.success(ctx, gin.H{"uploaded": successCount, "total": len(files)}, "success")
}

func (c *FileController) DownloadFile(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		c.fail(ctx, http.StatusBadRequest, "path is required")
		return
	}
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			c.fail(ctx, http.StatusInternalServerError, "failed to convert to absolute path: "+err.Error())
			return
		}
		path = absPath
	}
	fileOp := utils.NewFileOp()
	if !fileOp.Stat(path) {
		c.fail(ctx, http.StatusNotFound, "file not found")
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if info.IsDir() {
		c.fail(ctx, http.StatusBadRequest, "cannot download directory")
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename="+info.Name())
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	ctx.File(path)
}

func (c *FileController) handleJSON(ctx *gin.Context, fn func() error) {
	if err := fn(); err != nil {
		if strings.HasPrefix(err.Error(), "invalid character") || strings.HasPrefix(err.Error(), "unexpected EOF") || strings.Contains(err.Error(), "binding") {
			c.fail(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
			return
		}
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, nil, "success")
}
