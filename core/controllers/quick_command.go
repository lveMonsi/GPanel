package controllers

import (
	"gpanel/dto"
	"gpanel/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QuickCommandController struct {
	quickCommandService service.IQuickCommandService
}

func quickCommandError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

func quickCommandMessage(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": message})
}

func NewQuickCommandController() *QuickCommandController {
	return &QuickCommandController{
		quickCommandService: service.NewQuickCommandService(),
	}
}

// Create 创建快速命令
// POST /api/v1/quick-commands
func (c *QuickCommandController) Create(ctx *gin.Context) {
	var req dto.QuickCommandCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		quickCommandError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.quickCommandService.Create(req); err != nil {
		quickCommandError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	quickCommandMessage(ctx, "success")
}

// Update 更新快速命令
// POST /api/v1/quick-commands/update
func (c *QuickCommandController) Update(ctx *gin.Context) {
	var req dto.QuickCommandUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		quickCommandError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.quickCommandService.Update(req); err != nil {
		quickCommandError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	quickCommandMessage(ctx, "success")
}

// Delete 删除快速命令
// POST /api/v1/quick-commands/delete
func (c *QuickCommandController) Delete(ctx *gin.Context) {
	var req dto.QuickCommandDelete
	if err := ctx.ShouldBindJSON(&req); err != nil {
		quickCommandError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.quickCommandService.Delete(req); err != nil {
		quickCommandError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	quickCommandMessage(ctx, "success")
}

// Search 搜索快速命令
// POST /api/v1/quick-commands/search
func (c *QuickCommandController) Search(ctx *gin.Context) {
	var req dto.QuickCommandSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		quickCommandError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.quickCommandService.Search(req)
	if err != nil {
		quickCommandError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetAll 获取所有快速命令
// GET /api/v1/quick-commands/all
func (c *QuickCommandController) GetAll(ctx *gin.Context) {
	items, err := c.quickCommandService.GetAll()
	if err != nil {
		quickCommandError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, items)
}
