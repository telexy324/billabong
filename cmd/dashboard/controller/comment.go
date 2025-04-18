package controller

import (
	"encoding/json"
	"github.com/telexy324/billabong/pkg/markdown"
	"gorm.io/gorm"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get comment
// @Summary Get comment
// @Security BearerAuth
// @Schemes
// @Description Get comment
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.Comment]
// @Router /comment/{id} [get]
func getCommentById(c *gin.Context) (*model.Comment, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	uid := getUid(c)
	var comment model.Comment
	if err = singleton.DB.Where("id = ?", id).Find(&comment).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	comment.Content = markdown.ToHTML(comment.Content)
	comment.Favorited, err = singleton.FavoriteService.Exists(uid, model.EntityComment, comment.ID)
	if err != nil {
		return nil, err
	}
	comment.Liked, err = singleton.UserLikeService.Exists(uid, model.EntityComment, comment.ID)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// List comment
// @Summary List comment
// @Security BearerAuth
// @Schemes
// @Description List comment
// @Tags admin required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Comment]
// @Router /comment [get]
func listComment(c *gin.Context) ([]model.Comment, error) {
	idStr := c.Query("entityId")
	typeStr := c.Query("entityType")
	limitStr := c.Query("limit")
	var comments []model.Comment
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
	if limitStr != "" && limitStr != "0" {
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return nil, err
		}
		db = db.Limit(int(limit))
	}
	if err := db.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// Create comment
// @Summary Create comment
// @Security BearerAuth
// @Schemes
// @Description Create comment
// @Tags admin required
// @Accept json
// @param request body model.CommentForm true "Comment Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /comment [post]
func createComment(c *gin.Context) (uint64, error) {
	var tf model.CommentForm
	var t model.Comment
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
	t.CommentCount = tf.CommentCount
	t.Status = tf.Status
	t.Images = tf.Images
	if len(tf.Images) > 0 {
		if js, err := json.Marshal(tf.Images); err != nil {
			return 0, err
		} else {
			t.ImageList = string(js)
		}
	}

	if err := singleton.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&t).Error; err != nil {
			return err
		}
		// 更新点赞数
		if t.EntityType == model.EntityTopic {
			var oldTopic model.Topic
			db := tx.Where("id = ?", t.EntityId).Find(&oldTopic)
			upDateMap := make(map[string]interface{})
			upDateMap["CommentCount"] = oldTopic.CommentCount + 1
			upDateMap["LastCommentTime"] = t.CreatedAt
			upDateMap["LastCommentUserId"] = t.UserID
			txErr := db.Updates(upDateMap).Error
			if txErr != nil {
				return singleton.Localizer.ErrorT("update topic failed: %v", txErr)
			}
			return nil
		} else if t.EntityType == model.EntityComment {
			var oldComment model.Comment
			db := tx.Where("id = ?", t.EntityId).Find(&oldComment)
			upDateMap := make(map[string]interface{})
			upDateMap["CommentCount"] = oldComment.CommentCount + 1
			upDateMap["LastCommentTime"] = t.CreatedAt
			upDateMap["LastCommentUserId"] = t.UserID
			txErr := db.Updates(upDateMap).Error
			if txErr != nil {
				return singleton.Localizer.ErrorT("update topic failed: %v", txErr)
			}
			return nil
		} else {
			return singleton.Localizer.ErrorT("invalid entityType")
		}
	}); err != nil {
		return 0, err
	}

	return t.ID, nil
}

// Batch delete comments
// @Summary Batch delete comments
// @Security BearerAuth
// @Schemes
// @Description Batch delete comments
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/comment [post]
func batchDeleteComments(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	err := singleton.DB.Delete(&[]model.Comment{}, "id in ?", ids).Error
	return nil, err
}
