package controller

import (
	"database/sql"
	"encoding/json"
	"gorm.io/gorm"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get topic
// @Summary Get topic
// @Security BearerAuth
// @Schemes
// @Description Get topic
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.Topic]
// @Router /topic/{id} [get]
func getTopicById(c *gin.Context) (*model.Topic, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	uid := getUid(c)
	var topic model.Topic
	//if err = singleton.DB.Where("id = ?", id).Find(&topic).Error; err != nil {
	//	return nil, newGormError("%v", err)
	//}
	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", id).Find(&topic)
		txErr := db.Update("view_count", topic.ViewCount+1).Error
		if txErr != nil {
			return singleton.Localizer.ErrorT("update topic failed: %v", txErr)
		}
		return nil
	})
	formedTopic, err := singleton.TopicService.BuildTopic(topic, uid)
	if err != nil {
		return nil, err
	}
	return &formedTopic, nil
}

// Update password for current user
// @Summary Update password for current user
// @Security BearerAuth
// @Schemes
// @Description Update password for current user
// @Tags auth required
// @Accept json
// @param request body model.TopicForm true "Topic Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /topic/{id} [patch]
func updateTopic(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var tf model.TopicForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	var oldTopic model.Topic
	upDateMap := make(map[string]interface{})
	upDateMap["title"] = tf.Title
	upDateMap["content"] = tf.Content
	upDateMap["recommend"] = tf.Recommend
	upDateMap["sticky"] = tf.Sticky
	upDateMap["view_count"] = tf.ViewCount
	upDateMap["comment_count"] = tf.CommentCount
	upDateMap["like_count"] = tf.LikeCount
	upDateMap["status"] = tf.Status
	upDateMap["last_comment_user_id"] = tf.LastCommentUserId
	if len(tf.Affixes) > 0 {
		ids := make([]uint, 0, len(tf.Affixes))
		for _, image := range tf.Affixes {
			ids = append(ids, uint(image.ID))
		}
		if js, err := json.Marshal(ids); err != nil {
			return nil, err
		} else {
			upDateMap["imageList"] = string(js)
		}
	}

	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", id).Find(&oldTopic)
		txErr := db.Updates(upDateMap).Error
		if txErr != nil {
			return singleton.Localizer.ErrorT("update topic failed: %v", txErr)
		}
		return nil
	})
	return 0, err
}

// List topic
// @Summary List topic
// @Security BearerAuth
// @Schemes
// @Description List topic
// @Tags admin required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Topic]
// @Router /topic [get]
func listTopic(c *gin.Context) ([]model.Topic, error) {
	idStr := c.Query("groupId")
	var topics []model.Topic
	if idStr != "" {
		groupId, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		err = singleton.DB.Transaction(func(tx *gorm.DB) error {
			var tgt []model.TopicGroupTopic
			if err = tx.Where("topic_group_id = ?", groupId).Find(&tgt).Error; err != nil {
				return err
			}
			ids := make([]uint, 0, len(tgt))
			for _, t := range tgt {
				ids = append(ids, uint(t.TopicId))
			}
			if err = tx.Where("id IN (?)", ids).Find(&topics).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, newGormError("%v", err)
		}
	} else {
		if err := singleton.DB.Find(&topics).Error; err != nil {
			return nil, err
		}
	}
	var formedTopics []model.Topic
	for _, topic := range topics {
		formedTopic, err := singleton.TopicService.BuildTopic(topic, getUid(c))
		if err != nil {
			return nil, err
		}
		formedTopics = append(formedTopics, formedTopic)
	}
	return formedTopics, nil
}

// Create topic
// @Summary Create topic
// @Security BearerAuth
// @Schemes
// @Description Create topic
// @Tags admin required
// @Accept json
// @param request body model.TopicForm true "Topic Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /topic [post]
func createTopic(c *gin.Context) (uint64, error) {
	var tf model.TopicForm
	var t model.Topic
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	uid := getUid(c)

	t.UserID = uid
	t.Title = tf.Title
	t.Content = tf.Content
	t.Recommend = tf.Recommend
	t.RecommendTime = sql.NullTime{
		Time:  time.Unix(0, 0),
		Valid: true,
	}
	t.Sticky = tf.Sticky
	t.StickyTime = sql.NullTime{
		Time:  time.Unix(0, 0),
		Valid: true,
	}
	t.ViewCount = tf.ViewCount
	t.CommentCount = tf.CommentCount
	t.LikeCount = tf.LikeCount
	t.Status = tf.Status
	t.LastCommentTime = sql.NullTime{
		Time:  time.Unix(0, 0),
		Valid: true,
	}
	t.LastCommentUserId = tf.LastCommentUserId
	//if len(tf.Affixes) > 0 {
	//	if js, err := json.Marshal(tf.Affixes); err != nil {
	//		return 0, err
	//	} else {
	//		t.AffixList = string(js)
	//	}
	//}
	t.Affixes = tf.Affixes

	if tf.TopicGroup > 0 {
		var count int64
		singleton.DB.Model(&model.TopicGroup{}).Where("id = ?", tf.TopicGroup).Count(&count)
		if count <= 0 {
			return 0, singleton.Localizer.ErrorT("group not found")
		}
	}

	if err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&t).Error; err != nil {
			return err
		}
		var tg model.TopicGroupTopic
		tg.UserID = uid
		tg.TopicId = t.ID
		tg.TopicGroupId = tf.TopicGroup
		return tx.Create(&tg).Error
	}); err != nil {
		return 0, err
	}

	return t.ID, nil
}

// Batch delete topics
// @Summary Batch delete topics
// @Security BearerAuth
// @Schemes
// @Description Batch delete topics
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/topic [post]
func batchDeleteTopics(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	err := singleton.DB.Delete(&[]model.Topic{}, "id in ?", ids).Error
	return nil, err
}
