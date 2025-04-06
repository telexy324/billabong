package controller

import (
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// List topic group
// @Summary List topic group
// @Schemes
// @Description List topic group
// @Security BearerAuth
// @Tags common
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.TopicGroupResponseItem]
// @Router /topic-group [get]
func listTopicGroup(c *gin.Context) ([]*model.TopicGroupResponseItem, error) {
	var sg []model.TopicGroup
	if err := singleton.DB.Find(&sg).Error; err != nil {
		return nil, err
	}

	groupTopics := make(map[uint64][]uint64, 0)
	var sgs []model.TopicGroupTopic
	if err := singleton.DB.Find(&sgs).Error; err != nil {
		return nil, err
	}
	for _, s := range sgs {
		if _, ok := groupTopics[s.TopicGroupId]; !ok {
			groupTopics[s.TopicGroupId] = make([]uint64, 0)
		}
		groupTopics[s.TopicGroupId] = append(groupTopics[s.TopicGroupId], s.TopicId)
	}

	var sgRes []*model.TopicGroupResponseItem
	for _, s := range sg {
		sgRes = append(sgRes, &model.TopicGroupResponseItem{
			Group:  s,
			Topics: groupTopics[s.ID],
		})
	}

	return sgRes, nil
}

// New topic group
// @Summary New topic group
// @Schemes
// @Description New topic group
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param request body model.TopicGroupForm true "TopicGroupForm"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /topic-group [post]
func createTopicGroup(c *gin.Context) (uint64, error) {
	var sgf model.TopicGroupForm
	if err := c.ShouldBindJSON(&sgf); err != nil {
		return 0, err
	}
	sgf.Topics = slices.Compact(sgf.Topics)

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	uid := getUid(c)

	var sg model.TopicGroup
	sg.Name = sgf.Name
	sg.UserID = uid

	var count int64
	if err := singleton.DB.Model(&model.Topic{}).Where("id in (?)", sgf.Topics).Count(&count).Error; err != nil {
		return 0, newGormError("%v", err)
	}
	if count != int64(len(sgf.Topics)) {
		return 0, singleton.Localizer.ErrorT("have invalid topic id")
	}

	err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sg).Error; err != nil {
			return err
		}
		for _, s := range sgf.Topics {
			if err := tx.Create(&model.TopicGroupTopic{
				Common: model.Common{
					UserID: uid,
				},
				TopicGroupId: sg.ID,
				TopicId:      s,
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

// Edit topic group
// @Summary Edit topic group
// @Schemes
// @Description Edit topic group
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param id path uint true "ID"
// @Param body body model.TopicGroupForm true "TopicGroupForm"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /topic-group/{id} [patch]
func updateTopicGroup(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var sg model.TopicGroupForm
	if err := c.ShouldBindJSON(&sg); err != nil {
		return nil, err
	}
	sg.Topics = slices.Compact(sg.Topics)

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	var sgDB model.TopicGroup
	if err := singleton.DB.First(&sgDB, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("group id %d does not exist", id)
	}

	if !sgDB.HasPermission(c) {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	sgDB.Name = sg.Name

	var count int64
	if err := singleton.DB.Model(&model.Topic{}).Where("id in (?)", sg.Topics).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(sg.Topics)) {
		return nil, singleton.Localizer.ErrorT("have invalid topic id")
	}

	uid := getUid(c)

	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&sgDB).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&model.TopicGroupTopic{}, "topic_group_id = ?", id).Error; err != nil {
			return err
		}

		for _, s := range sg.Topics {
			if err := tx.Create(&model.TopicGroupTopic{
				Common: model.Common{
					UserID: uid,
				},
				TopicGroupId: sgDB.ID,
				TopicId:      s,
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

// Batch delete topic group
// @Summary Batch delete topic group
// @Security BearerAuth
// @Schemes
// @Description Batch delete topic group
// @Tags auth required
// @Accept json
// @param request body []uint64 true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/topic-group [post]
func batchDeleteTopicGroup(c *gin.Context) (any, error) {
	var sgs []uint64
	if err := c.ShouldBindJSON(&sgs); err != nil {
		return nil, err
	}

	var sg []model.TopicGroup
	if err := singleton.DB.Where("id in (?)", sgs).Find(&sg).Error; err != nil {
		return nil, err
	}

	for _, s := range sg {
		if !s.HasPermission(c) {
			return nil, singleton.Localizer.ErrorT("permission denied")
		}
	}

	err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&model.TopicGroup{}, "id in (?)", sgs).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&model.TopicGroupTopic{}, "topic_group_id in (?)", sgs).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, newGormError("%v", err)
	}

	return nil, nil
}
