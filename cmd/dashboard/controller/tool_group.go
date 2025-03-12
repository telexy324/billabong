package controller

import (
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// List tool group
// @Summary List tool group
// @Schemes
// @Description List tool group
// @Security BearerAuth
// @Tags common
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.ToolGroupResponseItem]
// @Router /tool-group [get]
func listToolGroup(c *gin.Context) ([]*model.ToolGroupResponseItem, error) {
	var sg []model.ToolGroup
	if err := singleton.DB.Find(&sg).Error; err != nil {
		return nil, err
	}

	groupTools := make(map[uint64][]uint64, 0)
	var sgs []model.ToolGroupTool
	if err := singleton.DB.Find(&sgs).Error; err != nil {
		return nil, err
	}
	for _, s := range sgs {
		if _, ok := groupTools[s.ToolGroupId]; !ok {
			groupTools[s.ToolGroupId] = make([]uint64, 0)
		}
		groupTools[s.ToolGroupId] = append(groupTools[s.ToolGroupId], s.ToolId)
	}

	var sgRes []*model.ToolGroupResponseItem
	for _, s := range sg {
		sgRes = append(sgRes, &model.ToolGroupResponseItem{
			Group: s,
			Tools: groupTools[s.ID],
		})
	}

	return sgRes, nil
}

// New tool group
// @Summary New tool group
// @Schemes
// @Description New tool group
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param body body model.ToolGroupForm true "ToolGroupForm"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /tool-group [post]
func createToolGroup(c *gin.Context) (uint64, error) {
	var sgf model.ToolGroupForm
	if err := c.ShouldBindJSON(&sgf); err != nil {
		return 0, err
	}
	sgf.Tools = slices.Compact(sgf.Tools)

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	uid := getUid(c)

	var sg model.ToolGroup
	sg.Name = sgf.Name
	sg.UserID = uid

	var count int64
	if err := singleton.DB.Model(&model.Tool{}).Where("id in (?)", sgf.Tools).Count(&count).Error; err != nil {
		return 0, newGormError("%v", err)
	}
	if count != int64(len(sgf.Tools)) {
		return 0, singleton.Localizer.ErrorT("have invalid tool id")
	}

	err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sg).Error; err != nil {
			return err
		}
		for _, s := range sgf.Tools {
			if err := tx.Create(&model.ToolGroupTool{
				Common: model.Common{
					UserID: uid,
				},
				ToolGroupId: sg.ID,
				ToolId:      s,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, newGormError("%v", err)
	}

	return sg.ID, nil
}

// Edit tool group
// @Summary Edit tool group
// @Schemes
// @Description Edit tool group
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param id path uint true "ID"
// @Param body body model.ToolGroupForm true "ToolGroupForm"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /tool-group/{id} [patch]
func updateToolGroup(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var sg model.ToolGroupForm
	if err := c.ShouldBindJSON(&sg); err != nil {
		return nil, err
	}
	sg.Tools = slices.Compact(sg.Tools)

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	var sgDB model.ToolGroup
	if err := singleton.DB.First(&sgDB, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("group id %d does not exist", id)
	}

	if !sgDB.HasPermission(c) {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	sgDB.Name = sg.Name

	var count int64
	if err := singleton.DB.Model(&model.Tool{}).Where("id in (?)", sg.Tools).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(sg.Tools)) {
		return nil, singleton.Localizer.ErrorT("have invalid tool id")
	}

	uid := getUid(c)

	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&sgDB).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&model.ToolGroupTool{}, "tool_group_id = ?", id).Error; err != nil {
			return err
		}

		for _, s := range sg.Tools {
			if err := tx.Create(&model.ToolGroupTool{
				Common: model.Common{
					UserID: uid,
				},
				ToolGroupId: sgDB.ID,
				ToolId:      s,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, newGormError("%v", err)
	}

	return nil, nil
}

// Batch delete tool group
// @Summary Batch delete tool group
// @Security BearerAuth
// @Schemes
// @Description Batch delete tool group
// @Tags auth required
// @Accept json
// @param request body []uint64 true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/tool-group [post]
func batchDeleteToolGroup(c *gin.Context) (any, error) {
	var sgs []uint64
	if err := c.ShouldBindJSON(&sgs); err != nil {
		return nil, err
	}

	var sg []model.ToolGroup
	if err := singleton.DB.Where("id in (?)", sgs).Find(&sg).Error; err != nil {
		return nil, err
	}

	for _, s := range sg {
		if !s.HasPermission(c) {
			return nil, singleton.Localizer.ErrorT("permission denied")
		}
	}

	err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&model.ToolGroup{}, "id in (?)", sgs).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&model.ToolGroupTool{}, "tool_group_id in (?)", sgs).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, newGormError("%v", err)
	}

	return nil, nil
}
