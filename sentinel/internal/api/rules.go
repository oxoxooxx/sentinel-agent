// Package api — 告警規則 API 處理器
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oxoxooxx/sentinel/internal/storage"
)

// RulesHandler 處理規則 CRUD 請求
type RulesHandler struct {
	db storage.DB
}

// NewRulesHandler 建立規則 handler
func NewRulesHandler(db storage.DB) *RulesHandler {
	return &RulesHandler{db: db}
}

// createRuleRequest 新增規則的請求 body
type createRuleRequest struct {
	Name        string `json:"name"      binding:"required"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`  // 指標型別以區分「未傳入」和「false」
	CondJSON    string `json:"cond_json" binding:"required"`
}

// updateRuleRequest 更新規則的請求 body
type updateRuleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	CondJSON    *string `json:"cond_json"`
}

// List 處理 GET /api/rules
func (h *RulesHandler) List(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "true"

	ctx := c.Request.Context()
	rules, err := h.db.ListRules(ctx, enabledOnly)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查詢規則失敗")
		return
	}

	if rules == nil {
		rules = []storage.Rule{}
	}

	ok(c, rules)
}

// Create 處理 POST /api/rules
func (h *RulesHandler) Create(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "請求格式錯誤: "+err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := storage.Rule{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     enabled,
		CondJSON:    req.CondJSON,
	}

	ctx := c.Request.Context()
	saved, err := h.db.SaveRule(ctx, rule)
	if err != nil {
		fail(c, http.StatusInternalServerError, "建立規則失敗")
		return
	}

	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: saved})
}

// Update 處理 PUT /api/rules/:id
func (h *RulesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	var req updateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "請求格式錯誤: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 先讀取現有規則
	rules, err := h.db.ListRules(ctx, false)
	if err != nil {
		fail(c, http.StatusInternalServerError, "讀取規則失敗")
		return
	}

	var existing *storage.Rule
	for _, r := range rules {
		if r.ID == id {
			copy := r
			existing = &copy
			break
		}
	}

	if existing == nil {
		fail(c, http.StatusNotFound, "找不到指定規則")
		return
	}

	// 套用部分更新（immutable pattern：建立新物件）
	updated := storage.Rule{
		ID:          existing.ID,
		Name:        existing.Name,
		Description: existing.Description,
		Enabled:     existing.Enabled,
		CondJSON:    existing.CondJSON,
		CreatedAt:   existing.CreatedAt,
	}

	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}
	if req.Enabled != nil {
		updated.Enabled = *req.Enabled
	}
	if req.CondJSON != nil {
		updated.CondJSON = *req.CondJSON
	}

	saved, err := h.db.SaveRule(ctx, updated)
	if err != nil {
		fail(c, http.StatusInternalServerError, "更新規則失敗")
		return
	}

	ok(c, saved)
}

// Delete 處理 DELETE /api/rules/:id
func (h *RulesHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "id 必須是整數")
		return
	}

	ctx := c.Request.Context()
	if err := h.db.DeleteRule(ctx, id); err != nil {
		fail(c, http.StatusInternalServerError, "刪除規則失敗")
		return
	}

	ok(c, gin.H{"deleted": id})
}
