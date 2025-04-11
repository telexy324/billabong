package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
	"strconv"
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
	var favorite *model.Favorite
	if err = singleton.DB.Where("id = ?", id).Find(favorite).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	return favorite, nil
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
	var favorites []model.Favorite
	if err := singleton.DB.Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

// Post user favorite
// @Summary Post user favorite
// @Security BearerAuth
// @Schemes
// @Description Post user favorite
// @Tags auth required
// @Accept json
// @param request body model.UserFavoriteForm true "Favorite Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /favorite [post]
func postFavorite(c *gin.Context) (any, error) {
	var tf model.UserFavoriteForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return nil, err
	}

	uid := getUid(c)

	if tf.EntityType == model.EntityTopic {
		return nil, singleton.FavoriteService.AddTopicFavorite(uid, tf.EntityId)
	} else {
		return nil, singleton.Localizer.ErrorT("entity unsupported")
	}
}

// Post user un favorite
// @Summary Post user un favorite
// @Security BearerAuth
// @Schemes
// @Description Post user un favorite
// @Tags auth required
// @Accept json
// @param request body model.UserFavoriteForm true "Favorite Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /unFavorite [post]
func postUnFavorite(c *gin.Context) (any, error) {
	var tf model.UserFavoriteForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return nil, err
	}

	if tf.EntityType == model.EntityTopic {
		return nil, singleton.FavoriteService.Delete(tf.EntityId)
	} else {
		return nil, singleton.Localizer.ErrorT("entity unsupported")
	}
}
