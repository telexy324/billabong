package controller

import (
	"encoding/json"
	"github.com/telexy324/billabong/pkg/markdown"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get favorite
// @Summary Get favorite
// @Security BearerAuth
// @Schemes
// @Description Get favorite
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.Favorite]
// @Router /favorite/{id} [get]
func getFavoriteById(c *gin.Context) (*model.Favorite, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var favorite model.Favorite
	if err = singleton.DB.Where("id = ?", id).Find(&favorite).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	favorite.Content = markdown.ToHTML(favorite.Content)
	return &favorite, nil
}

// List favorite
// @Summary List favorite
// @Security BearerAuth
// @Schemes
// @Description List favorite
// @Tags admin required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Favorite]
// @Router /favorite [get]
func listFavorite(c *gin.Context) ([]model.Favorite, error) {
	idStr := c.Query("entityId")
	typeStr := c.Query("entityType")
	var favorites []model.Favorite
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
	if err := db.Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

// Create favorite
// @Summary Create favorite
// @Security BearerAuth
// @Schemes
// @Description Create favorite
// @Tags admin required
// @Accept json
// @param request body model.FavoriteForm true "Favorite Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /favorite [post]
func createFavorite(c *gin.Context) (uint64, error) {
	var tf model.FavoriteForm
	var t model.Favorite
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
	t.FavoriteCount = tf.FavoriteCount
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

// Batch delete favorites
// @Summary Batch delete favorites
// @Security BearerAuth
// @Schemes
// @Description Batch delete favorites
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/favorite [post]
func batchDeleteFavorites(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	err := singleton.DB.Delete(&[]model.Favorite{}, "id in ?", ids).Error
	return nil, err
}
