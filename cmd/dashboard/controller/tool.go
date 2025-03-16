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
// @param request body model.ToolForm true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /tool/{id} [patch]
func updateTool(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var tf model.ToolForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	var oldTool model.Tool
	upDateMap := make(map[string]interface{})
	upDateMap["name"] = tf.Name
	upDateMap["summary"] = tf.Summary
	upDateMap["description"] = tf.Description
	upDateMap["disabled"] = tf.Disabled

	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", id).Find(&oldTool)
		if oldTool.Name != tf.Name {
			if !errors.Is(tx.Where("id <> ? AND name = ?", id, tf.Name).First(&model.Tool{}).Error, gorm.ErrRecordNotFound) {
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
// @param request body model.ToolForm true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /tool [post]
func createTool(c *gin.Context) (uint64, error) {
	var tf model.ToolForm
	var t model.Tool
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	//if tf.Name == "" {
	//	return 0, singleton.Localizer.ErrorT("tool name can't be empty")
	//}
	//if tf.Summary == "" {
	//	return 0, singleton.Localizer.ErrorT("tool summary can't be empty")
	//}
	t.Name = tf.Name
	t.Summary = tf.Summary
	t.Description = tf.Description
	t.Disabled = tf.Disabled

	if err := singleton.DB.Create(&t).Error; err != nil {
		return 0, err
	}

	return t.ID, nil
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
