package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CronjobController struct {
	svc *service.CronjobService
}

func NewCronjobController() *CronjobController {
	return &CronjobController{svc: service.NewCronjobService()}
}

func (c *CronjobController) ok(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

func (c *CronjobController) fail(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, gin.H{"code": status, "message": msg})
}

func (c *CronjobController) Create(ctx *gin.Context) {
	var req dto.CronjobCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.Create(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) Update(ctx *gin.Context) {
	var req dto.CronjobUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.Update(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) Delete(ctx *gin.Context) {
	var req dto.CronjobDelete
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.Delete(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) Search(ctx *gin.Context) {
	var req dto.CronjobSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	result, err := c.svc.Search(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, result)
}

func (c *CronjobController) Toggle(ctx *gin.Context) {
	var req dto.CronjobToggle
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.Toggle(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) HandleOnce(ctx *gin.Context) {
	var req dto.CronjobHandle
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.HandleOnce(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) StopRunning(ctx *gin.Context) {
	var req dto.CronjobStop
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.StopRunning(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) SearchRecords(ctx *gin.Context) {
	var req dto.RecordSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	result, err := c.svc.SearchRecords(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, result)
}

func (c *CronjobController) GetRecordLog(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.fail(ctx, http.StatusBadRequest, "invalid record id")
		return
	}
	content, err := c.svc.GetRecordLog(uint(id))
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, content)
}

func (c *CronjobController) CleanRecords(ctx *gin.Context) {
	var req dto.RecordClean
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.svc.CleanRecords(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

func (c *CronjobController) GetNextExecTimes(ctx *gin.Context) {
	var req dto.NextTimesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	times, err := c.svc.GetNextExecTimes(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, gin.H{"times": times})
}
