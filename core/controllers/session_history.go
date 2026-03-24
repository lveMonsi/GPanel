package controllers

import (
	"gpanel/dto"
	"gpanel/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var sessionHistoryService service.ISessionHistoryService

func sessionHistoryError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

func sessionHistoryMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": message})
}

// InitSessionHistoryController 初始化会话历史控制器
func InitSessionHistoryController() {
	sessionHistoryService = service.NewSessionHistoryService()
}

// CreateSessionHistory 创建会话历史
// @Summary 创建会话历史
// @Tags SessionHistory
// @Accept json
// @Produce json
// @Param request body dto.SessionHistoryCreate true "会话信息"
// @Success 200 {object} models.SessionHistory
// @Router /api/v1/session-histories [post]
func CreateSessionHistory(c *gin.Context) {
	var req dto.SessionHistoryCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		sessionHistoryError(c, http.StatusBadRequest, err.Error())
		return
	}

	history, err := sessionHistoryService.Create(req)
	if err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, history)
}

// UpdateSessionHistory 更新会话历史
// @Summary 更新会话历史
// @Tags SessionHistory
// @Accept json
// @Produce json
// @Param request body dto.SessionHistoryUpdate true "会话信息"
// @Success 200 {object} map[string]string
// @Router /api/v1/session-histories [put]
func UpdateSessionHistory(c *gin.Context) {
	var req dto.SessionHistoryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		sessionHistoryError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := sessionHistoryService.Update(req); err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	sessionHistoryMessage(c, "Session history updated")
}

// GetSessionHistoryByID 获取会话历史详情
// @Summary 获取会话历史详情
// @Tags SessionHistory
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} dto.SessionHistoryInfo
// @Router /api/v1/session-histories/{id} [get]
func GetSessionHistoryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		sessionHistoryError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	history, err := sessionHistoryService.GetByID(uint(id))
	if err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, history)
}

// ListSessionHistories 获取会话历史列表
// @Summary 获取会话历史列表
// @Tags SessionHistory
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/session-histories [get]
func ListSessionHistories(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 10)

	histories, total, err := sessionHistoryService.List(page, pageSize)
	if err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  histories,
		"total": total,
	})
}

// SearchSessionHistories 搜索会话历史
// @Summary 搜索会话历史
// @Tags SessionHistory
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param hostID query int false "主机ID"
// @Param hostAddr query string false "主机地址"
// @Param userName query string false "用户名"
// @Param startDate query int false "开始日期"
// @Param endDate query int false "结束日期"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/session-histories/search [post]
func SearchSessionHistories(c *gin.Context) {
	var req dto.SessionHistorySearch
	if err := c.ShouldBindJSON(&req); err != nil {
		sessionHistoryError(c, http.StatusBadRequest, err.Error())
		return
	}

	histories, total, err := sessionHistoryService.Search(req)
	if err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  histories,
		"total": total,
	})
}

// DeleteSessionHistory 删除会话历史
// @Summary 删除会话历史
// @Tags SessionHistory
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/session-histories/{id} [delete]
func DeleteSessionHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		sessionHistoryError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := sessionHistoryService.Delete(uint(id)); err != nil {
		sessionHistoryError(c, http.StatusInternalServerError, err.Error())
		return
	}

	sessionHistoryMessage(c, "Session history deleted")
}
