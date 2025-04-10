package controller

import (
	"encoding/json"
	"github.com/telexy324/billabong/pkg/markdown"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get like
// @Summary Get like
// @Security BearerAuth
// @Schemes
// @Description Get like
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.Like]
// @Router /like/{id} [get]
func getLikeById(c *gin.Context) (*model.Like, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var like model.Like
	if err = singleton.DB.Where("id = ?", id).Find(&like).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	like.Content = markdown.ToHTML(like.Content)
	return &like, nil
}

// List like
// @Summary List like
// @Security BearerAuth
// @Schemes
// @Description List like
// @Tags admin required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Like]
// @Router /like [get]
func listLike(c *gin.Context) ([]model.Like, error) {
	idStr := c.Query("entityId")
	typeStr := c.Query("entityType")
	var likes []model.Like
	db := singleton.DB
	if idStr != "" && typeStr == "" {
		return nil, singleton.Localizer.ErrorT("should have entity type")
	}
	if typeStr != "" {
		entityType, err := strconv.ParseInt(typeStr, 10, 64)
		if err != nil {
			return nil, err
		}
		if idStr != "" {
			entityId, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return nil, err
			}
			db = db.Where("entity_type = ? and entity_id = ?", entityType, entityId)
		} else {
			db = db.Where("entity_type = ?", entityType)
		}
	}
	if err := db.Find(&likes).Error; err != nil {
		return nil, err
	}
	return likes, nil
}

// Create like
// @Summary Create like
// @Security BearerAuth
// @Schemes
// @Description Create like
// @Tags admin required
// @Accept json
// @param request body model.LikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /like [post]
func createLike(c *gin.Context) (uint64, error) {
	var tf model.LikeForm
	var t model.Like
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return 0, singleton.Localizer.ErrorT("unauthorized")
	}

	t.EntityType = tf.EntityType
	t.EntityId = tf.EntityId
	t.Content = tf.Content
	t.ContentType = tf.ContentType
	t.QuoteId = tf.QuoteId
	t.LikeCount = tf.LikeCount
	t.LikeCount = tf.LikeCount
	t.Status = tf.Status
	t.Images = tf.Images
	if len(tf.Images) > 0 {
		if js, err := json.Marshal(tf.Images); err != nil {
			return 0, err
		} else {
			t.ImageList = string(js)
		}
	}

	if err := singleton.DB.Create(&t).Error; err != nil {
		return 0, err
	}

	return t.ID, nil
}

// Batch delete likes
// @Summary Batch delete likes
// @Security BearerAuth
// @Schemes
// @Description Batch delete likes
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/like [post]
func batchDeleteLikes(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	err := singleton.DB.Delete(&[]model.Like{}, "id in ?", ids).Error
	return nil, err
}

// Post user like
// @Summary Post user like
// @Security BearerAuth
// @Schemes
// @Description Post user like
// @Tags auth required
// @Accept json
// @param request body model.LikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /like [post]
func postLike(c *gin.Context) error {
	var tf model.UserLikeForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return err
	}

	uid := getUid(c)

	if tf.EntityType == model.EntityComment {
		return singleton.UserLikeService.CommentLike(uid, tf.EntityId)
	} else if tf.EntityType == model.EntityTopic {
		return singleton.UserLikeService.TopicLike(uid, tf.EntityId)
	} else {
		return singleton.Localizer.ErrorT("entity unsupported")
	}
}

// Post user un like
// @Summary Post user un like
// @Security BearerAuth
// @Schemes
// @Description Post user un like
// @Tags auth required
// @Accept json
// @param request body model.LikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /unlike [post]
func postUnLike(c *gin.Context) error {
	var tf model.UserLikeForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return err
	}

	uid := getUid(c)

	if tf.EntityType == model.EntityComment {
		return singleton.UserLikeService.CommentUnLike(uid, tf.EntityId)
	} else if tf.EntityType == model.EntityTopic {
		return singleton.UserLikeService.TopicUnLike(uid, tf.EntityId)
	} else {
		return singleton.Localizer.ErrorT("entity unsupported")
	}
}
