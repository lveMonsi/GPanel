package controllers

import (
	"io"
	"net/http"
	"net/url"

	"gpanel/dto"
	"gpanel/utils"

	"github.com/gin-gonic/gin"
)

// FileController 文件控制器（代理到Agent）
type FileController struct {
	agentClient *utils.AgentClient
}

func fileError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

// NewFileController 创建文件控制器
func NewFileController() *FileController {
	client, _ := utils.NewAgentClient()
	return &FileController{agentClient: client}
}

func (c *FileController) ensureClient() error {
	if c.agentClient != nil {
		return nil
	}
	client, err := utils.NewAgentClient()
	if err != nil {
		return err
	}
	c.agentClient = client
	return nil
}

func (c *FileController) GetDrives(ctx *gin.Context) {
	c.proxyGet(ctx, "/api/v1/files/drives")
}

func (c *FileController) ListFiles(ctx *gin.Context) {
	var req dto.FileListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/list", req)
}

func (c *FileController) CreateFile(ctx *gin.Context) {
	var req dto.FileCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/create", req)
}

func (c *FileController) DeleteFile(ctx *gin.Context) {
	var req dto.FileDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/delete", req)
}

func (c *FileController) RenameFile(ctx *gin.Context) {
	var req dto.FileRenameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/rename", req)
}

func (c *FileController) MoveFile(ctx *gin.Context) {
	var req dto.FileMoveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/move", req)
}

func (c *FileController) GetFileContent(ctx *gin.Context) {
	var req dto.FileContentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/content", req)
}

func (c *FileController) SaveFileContent(ctx *gin.Context) {
	var req dto.FileEditReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/save", req)
}

func (c *FileController) UploadFile(ctx *gin.Context) {
	if err := c.ensureClient(); err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	resp, err := c.agentClient.StreamRequest(http.MethodPost, "/api/v1/files/upload", ctx.Request.Body, ctx.GetHeader("Content-Type"))
	if err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (c *FileController) DownloadFile(ctx *gin.Context) {
	if err := c.ensureClient(); err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	path := url.QueryEscape(ctx.Query("path"))
	resp, err := c.agentClient.StreamRequest(http.MethodGet, "/api/v1/files/download?path="+path, nil, "")
	if err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			ctx.Writer.Header().Add(key, value)
		}
	}
	ctx.Status(resp.StatusCode)
	_, _ = io.Copy(ctx.Writer, resp.Body)
}

func (c *FileController) GetDirSize(ctx *gin.Context) {
	var req dto.DirSizeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/size", req)
}

func (c *FileController) CompressFiles(ctx *gin.Context) {
	var req dto.FileCompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/compress", req)
}

func (c *FileController) DecompressFile(ctx *gin.Context) {
	var req dto.FileDecompressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/decompress", req)
}

func (c *FileController) ChmodFile(ctx *gin.Context) {
	var req dto.FileChmodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/chmod", req)
}

func (c *FileController) ChownFile(ctx *gin.Context) {
	var req dto.FileChownReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/chown", req)
}

func (c *FileController) GetProgress(ctx *gin.Context) {
	key := url.QueryEscape(ctx.Query("key"))
	c.proxyGet(ctx, "/api/v1/files/progress?key="+key)
}

func (c *FileController) PreviewFile(ctx *gin.Context) {
	var req dto.FilePreviewReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/preview", req)
}

func (c *FileController) RemoteDownload(ctx *gin.Context) {
	var req dto.RemoteDownloadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fileError(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	c.proxyJSON(ctx, "/api/v1/files/remote-download", req)
}

func (c *FileController) proxyGet(ctx *gin.Context, path string) {
	if err := c.ensureClient(); err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodGet, path, nil)
	if err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

func (c *FileController) proxyJSON(ctx *gin.Context, path string, req interface{}) {
	if err := c.ensureClient(); err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, path, req)
	if err != nil {
		fileError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}
