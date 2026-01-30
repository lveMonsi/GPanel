package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FirewallController 防火墙控制器
type FirewallController struct {
	firewallService *service.FirewallService
}

// NewFirewallController 创建防火墙控制器
func NewFirewallController() (*FirewallController, error) {
	return &FirewallController{
		firewallService: service.NewFirewallService(),
	}, nil
}

// LoadBaseInfo 加载防火墙基础信息
func (c *FirewallController) LoadBaseInfo(ctx *gin.Context) {
	info, err := c.firewallService.LoadBaseInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": info})
}

// SearchRules 搜索防火墙规则
func (c *FirewallController) SearchRules(ctx *gin.Context) {
	var req dto.RuleSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := c.firewallService.SearchRules(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// OperateFirewall 操作防火墙
func (c *FirewallController) OperateFirewall(ctx *gin.Context) {
	var req dto.FirewallOperation
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateFirewall(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// OperatePortRule 操作端口规则
func (c *FirewallController) OperatePortRule(ctx *gin.Context) {
	var req dto.PortRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperatePortRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// UpdatePortRule 更新端口规则
func (c *FirewallController) UpdatePortRule(ctx *gin.Context) {
	var req dto.PortRuleUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.UpdatePortRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// OperateIPRule 操作IP规则
func (c *FirewallController) OperateIPRule(ctx *gin.Context) {
	var req dto.IPRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateIPRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// UpdateIPRule 更新IP规则
func (c *FirewallController) UpdateIPRule(ctx *gin.Context) {
	var req dto.IPRuleUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.UpdateIPRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// OperateForwardRule 操作端口转发规则
func (c *FirewallController) OperateForwardRule(ctx *gin.Context) {
	var req dto.ForwardRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateForwardRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}