package controller

import (
	"errors"
	"gorm.io/gorm"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get tool
// @Summary Get tool
// @Security BearerAuth
// @Schemes
// @Description Get tool
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.Tool]
// @Router /tool/{id} [get]
func getToolById(c *gin.Context) (*model.Tool, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var tool *model.Tool
	if err = singleton.DB.Where("id = ?", id).Find(tool).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	return tool, nil
}

// Update password for current user
// @Summary Update password for current user
// @Security BearerAuth
// @Schemes
// @Description Update password for current user
// @Tags auth required
// @Accept json
// @param request body model.Tool true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /updateTool [post]
func updateTool(c *gin.Context) (any, error) {
	var pf model.Tool
	if err := c.ShouldBindJSON(&pf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	var oldTool model.Tool
	upDateMap := make(map[string]interface{})
	upDateMap["name"] = pf.Name
	upDateMap["summary"] = pf.Summary
	upDateMap["description"] = pf.Description
	upDateMap["downloads"] = pf.Downloads
	upDateMap["disabled"] = pf.Disabled

	err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", pf.ID).Find(&oldTool)
		if oldTool.Name != pf.Name {
			if !errors.Is(tx.Where("id <> ? AND name = ?", pf.ID, pf.Name).First(&model.Tool{}).Error, gorm.ErrRecordNotFound) {
				return singleton.Localizer.ErrorT("update tool failed: same name exists")
			}
		}
		txErr := db.Updates(upDateMap).Error
		if txErr != nil {
			return singleton.Localizer.ErrorT("update tool failed: %v", txErr)
		}
		return nil
	})
	return 0, err
}

// List tool
// @Summary List tool
// @Security BearerAuth
// @Schemes
// @Description List tool
// @Tags admin required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Tool]
// @Router /tool [get]
func listTool(c *gin.Context) ([]model.Tool, error) {
	var tools []model.Tool
	if err := singleton.DB.Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

// Create tool
// @Summary Create tool
// @Security BearerAuth
// @Schemes
// @Description Create tool
// @Tags admin required
// @Accept json
// @param request body model.Tool true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /tool [post]
func createTool(c *gin.Context) (uint64, error) {
	var uf model.Tool
	if err := c.ShouldBindJSON(&uf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	if uf.Name == "" {
		return 0, singleton.Localizer.ErrorT("tool name can't be empty")
	}
	if uf.Summary == "" {
		return 0, singleton.Localizer.ErrorT("tool summary can't be empty")
	}

	if err := singleton.DB.Create(&uf).Error; err != nil {
		return 0, err
	}

	return uf.ID, nil
}

// Batch delete tools
// @Summary Batch delete tools
// @Security BearerAuth
// @Schemes
// @Description Batch delete tools
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/tool [post]
func batchDeleteTools(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	err := singleton.DB.Delete(&[]model.Tool{}, "id in ?", ids).Error
	return nil, err
}
