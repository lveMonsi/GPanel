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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.quickCommandService.Create(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Update 更新快速命令
// POST /api/v1/quick-commands/update
func (c *QuickCommandController) Update(ctx *gin.Context) {
	var req dto.QuickCommandUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.quickCommandService.Update(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Delete 删除快速命令
// POST /api/v1/quick-commands/delete
func (c *QuickCommandController) Delete(ctx *gin.Context) {
	var req dto.QuickCommandDelete
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.quickCommandService.Delete(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Search 搜索快速命令
// POST /api/v1/quick-commands/search
func (c *QuickCommandController) Search(ctx *gin.Context) {
	var req dto.QuickCommandSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := c.quickCommandService.Search(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetAll 获取所有快速命令
// GET /api/v1/quick-commands/all
func (c *QuickCommandController) GetAll(ctx *gin.Context) {
	items, err := c.quickCommandService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, items)
}