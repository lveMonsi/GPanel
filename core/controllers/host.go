package controllers

import (
	"fmt"
	"gpanel/dto"
	"gpanel/service"
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var hostService service.IHostService

func hostError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

func hostMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": message})
}

// InitHostController 初始化主机控制器
func InitHostController() {
	hostService = service.NewHostService()
}

// HostGroup operations

// CreateGroup 创建主机分组
// @Summary 创建主机分组
// @Tags HostGroup
// @Accept json
// @Produce json
// @Param request body dto.HostGroupOperate true "分组信息"
// @Success 200 {object} models.HostGroup
// @Router /api/v1/host-groups [post]
func CreateGroup(c *gin.Context) {
	var req dto.HostGroupOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	group, err := hostService.CreateGroup(req)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateGroup 更新主机分组
// @Summary 更新主机分组
// @Tags HostGroup
// @Accept json
// @Produce json
// @Param id path int true "分组ID"
// @Param request body dto.HostGroupOperate true "分组信息"
// @Success 200 {object} map[string]string
// @Router /api/v1/host-groups/{id} [put]
func UpdateGroup(c *gin.Context) {
	var req dto.HostGroupOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}
	req.ID = id

	if err := hostService.UpdateGroup(req); err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	hostMessage(c, "Group updated")
}

// DeleteGroup 删除主机分组
// @Summary 删除主机分组
// @Tags HostGroup
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/host-groups/{id} [delete]
func DeleteGroup(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := hostService.DeleteGroup(id); err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	hostMessage(c, "Group deleted")
}

// GetGroupByID 获取主机分组详情
// @Summary 获取主机分组详情
// @Tags HostGroup
// @Produce json
// @Param id path int true "分组ID"
// @Success 200 {object} dto.HostGroupInfo
// @Router /api/v1/host-groups/{id} [get]
func GetGroupByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	group, err := hostService.GetGroupByID(id)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, group)
}

// ListGroups 获取主机分组列表
// @Summary 获取主机分组列表
// @Tags HostGroup
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param info query string false "搜索关键词"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/host-groups [get]
func ListGroups(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 10)
	info := c.Query("info")

	req := dto.HostGroupSearch{
		PageInfo: dto.PageInfo{
			Page:     page,
			PageSize: pageSize,
		},
		Info: info,
	}

	groups, total, err := hostService.ListGroups(req)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  groups,
		"total": total,
	})
}

// GetHostTree 获取主机树
// @Summary 获取主机树
// @Tags Host
// @Produce json
// @Success 200 {array} dto.HostTreeNode
// @Router /api/v1/hosts/tree [get]
func GetHostTree(c *gin.Context) {
	tree, err := hostService.GetHostTree()
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tree)
}

// Host operations

// CreateHost 创建主机
// @Summary 创建主机
// @Tags Host
// @Accept json
// @Produce json
// @Param request body dto.HostOperate true "主机信息"
// @Success 200 {object} dto.HostInfo
// @Router /api/v1/hosts [post]
func CreateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	host, err := hostService.CreateHost(req)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, host)
}

// UpdateHost 更新主机
// @Summary 更新主机
// @Tags Host
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Param request body dto.HostOperate true "主机信息"
// @Success 200 {object} map[string]string
// @Router /api/v1/hosts/{id} [put]
func UpdateHost(c *gin.Context) {
	var req dto.HostOperate
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}
	req.ID = id

	if err := hostService.UpdateHost(req); err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	hostMessage(c, "Host updated")
}

// DeleteHost 删除主机
// @Summary 删除主机
// @Tags Host
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/hosts/{id} [delete]
func DeleteHost(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := hostService.DeleteHost(id); err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	hostMessage(c, "Host deleted")
}

// GetHostByID 获取主机详情
// @Summary 获取主机详情
// @Tags Host
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} dto.HostInfo
// @Router /api/v1/hosts/{id} [get]
func GetHostByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	host, err := hostService.GetHostByID(id)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, host)
}

// ListHosts 获取主机列表
// @Summary 获取主机列表
// @Tags Host
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param groupID query int false "分组ID"
// @Param info query string false "搜索关键词"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/hosts [get]
func ListHosts(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 10)
	groupID := parseIntQuery(c, "groupID", 0)
	info := c.Query("info")

	req := dto.HostSearch{
		PageInfo: dto.PageInfo{
			Page:     page,
			PageSize: pageSize,
		},
		GroupID: uint(groupID),
		Info:    info,
	}

	hosts, total, err := hostService.ListHosts(req)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  hosts,
		"total": total,
	})
}

// TestHostConnection 测试主机连接
// @Summary 测试主机连接
// @Tags Host
// @Accept json
// @Produce json
// @Param request body dto.HostConnTest true "连接信息"
// @Success 200 {object} map[string]string
// @Router /api/v1/hosts/test [post]
func TestHostConnection(c *gin.Context) {
	var req dto.HostConnTest
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}
	agentClient, err := utils.NewAgentClient()
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp, statusCode, err := agentClient.RequestWithStatus(http.MethodPost, "/api/v1/hosts/test", gin.H{
		"addr":       req.Addr,
		"port":       req.Port,
		"user":       req.User,
		"authMode":   req.AuthMode,
		"password":   req.Password,
		"privateKey": req.PrivateKey,
		"passPhrase": req.PassPhrase,
	})
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(statusCode, "application/json", resp)
}

// MoveHosts 移动主机到其他分组
// @Summary 移动主机到其他分组
// @Tags Host
// @Accept json
// @Produce json
// @Param request body dto.HostMove true "移动信息"
// @Success 200 {object} map[string]string
// @Router /api/v1/hosts/move [post]
func MoveHosts(c *gin.Context) {
	var req dto.HostMove
	if err := c.ShouldBindJSON(&req); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := hostService.MoveHosts(req); err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	hostMessage(c, "Hosts moved")
}

// GetHostForTerminal 获取主机连接信息（用于终端连接）
// @Summary 获取主机连接信息
// @Tags Host
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Host
// @Router /api/v1/hosts/{id}/connection [get]
func GetHostForTerminal(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	host, err := hostService.GetHostForConnection(id)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, host)
}

// ExportHosts 导出主机列表
// @Summary 导出主机列表
// @Tags Host
// @Produce json
// @Param encrypted query bool false "是否加密导出敏感信息" default(true)
// @Success 200 {array} dto.HostOperate
// @Router /api/v1/hosts/export [get]
func ExportHosts(c *gin.Context) {
	// 默认加密导出
	encrypted := true
	if encryptedStr := c.Query("encrypted"); encryptedStr == "false" {
		encrypted = false
	}

	hosts, err := hostService.ExportHosts(encrypted)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, hosts)
}

// ImportHosts 导入主机列表
// @Summary 导入主机列表
// @Tags Host
// @Accept json
// @Produce json
// @Param request body []dto.HostOperate true "主机列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/hosts/import [post]
func ImportHosts(c *gin.Context) {
	var hosts []dto.HostOperate
	if err := c.ShouldBindJSON(&hosts); err != nil {
		hostError(c, http.StatusBadRequest, err.Error())
		return
	}

	success, fail, err := hostService.ImportHosts(hosts)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"success": success,
		"fail":    fail,
		"message": fmt.Sprintf("成功导入 %d 个主机，失败 %d 个", success, fail),
	})
}

// GetHostForConnection 获取主机连接信息（用于终端连接）
// @Summary 获取主机连接信息
// @Tags Host
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} dto.HostConnInfo
// @Router /api/v1/hosts/{id}/connection [get]
func GetHostForConnection(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		hostError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	host, err := hostService.GetHostForConnection(id)
	if err != nil {
		hostError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, host)
}

// Helper functions

func parseUintParam(c *gin.Context, key string) (uint, error) {
	value := c.Param(key)
	var result uint
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
