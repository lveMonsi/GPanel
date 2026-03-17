package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gpanel/dto"
	"gpanel/service"
	"gpanel/utils"

	"github.com/gin-gonic/gin"
)

// FileController 文件控制器
type FileController struct {
	fileService *service.FileService
}

// NewFileController 创建文件控制器
func NewFileController() *FileController {
	return &FileController{
		fileService: service.NewFileService(),
	}
}

// GetDrives 获取系统盘符列表
// @Summary 获取系统盘符列表
// @Tags File
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/drives [get]
func (c *FileController) GetDrives(ctx *gin.Context) {
	var drives []string

	if runtime.GOOS == "windows" {
		// Windows 系统：获取所有可用的盘符
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drivePath := string(drive) + ":\\"
			_, err := os.Stat(drivePath)
			if err == nil {
				drives = append(drives, drivePath)
			}
		}
	} else {
		// Linux/Unix 系统：根目录
		drives = append(drives, "/")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    drives,
	})
}

// ListFiles 获取文件列表
// @Summary 获取文件列表
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileListReq true "文件列表请求"
// @Success 200 {object} dto.FileListRes
// @Router /api/v1/files/list [post]
func (c *FileController) ListFiles(ctx *gin.Context) {
	var req dto.FileListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := c.fileService.GetFileList(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// CreateFile 创建文件或目录
// @Summary 创建文件或目录
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileCreateReq true "创建文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/create [post]
func (c *FileController) CreateFile(ctx *gin.Context) {
	var req dto.FileCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.CreateFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// DeleteFile 删除文件或目录
// @Summary 删除文件或目录
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileDeleteReq true "删除文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/delete [post]
func (c *FileController) DeleteFile(ctx *gin.Context) {
	var req dto.FileDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.DeleteFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// RenameFile 重命名文件或目录
// @Summary 重命名文件或目录
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileRenameReq true "重命名文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/rename [post]
func (c *FileController) RenameFile(ctx *gin.Context) {
	var req dto.FileRenameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.RenameFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// MoveFile 移动或复制文件
// @Summary 移动或复制文件
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileMoveReq true "移动文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/move [post]
func (c *FileController) MoveFile(ctx *gin.Context) {
	var req dto.FileMoveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.MoveFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetFileContent 获取文件内容
// @Summary 获取文件内容
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileContentReq true "获取文件内容请求"
// @Success 200 {object} dto.FileContentRes
// @Router /api/v1/files/content [post]
func (c *FileController) GetFileContent(ctx *gin.Context) {
	var req dto.FileContentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := c.fileService.GetFileContent(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// SaveFileContent 保存文件内容
// @Summary 保存文件内容
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileEditReq true "保存文件内容请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/save [post]
func (c *FileController) SaveFileContent(ctx *gin.Context) {
	var req dto.FileEditReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.SaveFileContent(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// UploadFile 上传文件
// @Summary 上传文件
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param path formData string true "目标路径"
// @Param file formData file true "文件"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/upload [post]
func (c *FileController) UploadFile(ctx *gin.Context) {
	path := ctx.PostForm("path")
	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "path is required",
		})
		return
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "failed to convert to absolute path: " + err.Error(),
			})
			return
		}
		path = absPath
	}

	overwrite := ctx.PostForm("overwrite")
	overwriteFlag := strings.ToLower(overwrite) == "true"

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(path) {
		if err := fileOp.CreateDir(path, 0755); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "failed to create directory: " + err.Error(),
			})
			return
		}
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "failed to parse form: " + err.Error(),
		})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "no files uploaded",
		})
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
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			continue
		}
		defer dstFile.Close()

		if _, err := dstFile.ReadFrom(srcFile); err != nil {
			continue
		}

		successCount++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"uploaded": successCount,
			"total":    len(files),
		},
	})
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Tags File
// @Produce application/octet-stream
// @Param path query string true "文件路径"
// @Router /api/v1/files/download [get]
func (c *FileController) DownloadFile(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "path is required",
		})
		return
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "failed to convert to absolute path: " + err.Error(),
			})
			return
		}
		path = absPath
	}

	fileOp := utils.NewFileOp()
	if !fileOp.Stat(path) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "file not found",
		})
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	if info.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "cannot download directory",
		})
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename="+info.Name())
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(info.Size(), 10))

	ctx.File(path)
}

// GetDirSize 获取目录大小
// @Summary 获取目录大小
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.DirSizeReq true "获取目录大小请求"
// @Success 200 {object} dto.DirSizeRes
// @Router /api/v1/files/size [post]
func (c *FileController) GetDirSize(ctx *gin.Context) {
	var req dto.DirSizeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := c.fileService.GetDirSize(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// CompressFiles 压缩文件
// @Summary 压缩文件
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileCompressReq true "压缩文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/compress [post]
func (c *FileController) CompressFiles(ctx *gin.Context) {
	var req dto.FileCompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	progressKey, err := c.fileService.CompressFiles(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"key": progressKey,
		},
	})
}

// DecompressFile 解压文件
// @Summary 解压文件
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileDecompressReq true "解压文件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/decompress [post]
func (c *FileController) DecompressFile(ctx *gin.Context) {
	var req dto.FileDecompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	progressKey, err := c.fileService.DecompressFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"key": progressKey,
		},
	})
}

// ChmodFile 修改文件权限
// @Summary 修改文件权限
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileChmodReq true "修改文件权限请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/chmod [post]
func (c *FileController) ChmodFile(ctx *gin.Context) {
	var req dto.FileChmodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.ChmodFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// ChownFile 修改文件所有者
// @Summary 修改文件所有者
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FileChownReq true "修改文件所有者请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/chown [post]
func (c *FileController) ChownFile(ctx *gin.Context) {
	var req dto.FileChownReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.fileService.ChownFile(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// GetProgress 获取操作进度
// @Summary 获取操作进度
// @Tags File
// @Accept json
// @Produce json
// @Param key query string true "进度键"
// @Success 200 {object} dto.ProgressInfo
// @Router /api/v1/files/progress [get]
func (c *FileController) GetProgress(ctx *gin.Context) {
	key := ctx.Query("key")
	if key == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "key is required",
		})
		return
	}

	progress := c.fileService.GetProgress(key)
	if progress == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "progress not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    progress,
	})
}

// PreviewFile 预览文件
// @Summary 预览文件
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.FilePreviewReq true "预览文件请求"
// @Success 200 {object} dto.FilePreviewRes
// @Router /api/v1/files/preview [post]
func (c *FileController) PreviewFile(ctx *gin.Context) {
	var req dto.FilePreviewReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := c.fileService.PreviewFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// RemoteDownload 远程下载文件
// @Summary 远程下载文件
// @Tags File
// @Accept json
// @Produce json
// @Param request body dto.RemoteDownloadReq true "远程下载请求"
// @Success 200 {object} dto.RemoteDownloadRes
// @Router /api/v1/files/remote-download [post]
func (c *FileController) RemoteDownload(ctx *gin.Context) {
	var req dto.RemoteDownloadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := c.fileService.RemoteDownload(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}