package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Post user like
// @Summary Post user like
// @Security BearerAuth
// @Schemes
// @Description Post user like
// @Tags auth required
// @Accept json
// @param request body model.UserLikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[int]
// @Router /like [post]
func postLike(c *gin.Context) (int64, error) {
	var tf model.UserLikeForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	uid := getUid(c)

	if tf.EntityType == model.EntityComment {
		return singleton.UserLikeService.CommentLike(uid, tf.EntityId)
	} else if tf.EntityType == model.EntityTopic {
		return singleton.UserLikeService.TopicLike(uid, tf.EntityId)
	} else {
		return 0, singleton.Localizer.ErrorT("entity unsupported")
	}
}

// Post user un like
// @Summary Post user un like
// @Security BearerAuth
// @Schemes
// @Description Post user un like
// @Tags auth required
// @Accept json
// @param request body model.UserLikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[int]
// @Router /unlike [post]
func postUnLike(c *gin.Context) (int64, error) {
	var tf model.UserLikeForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return 0, err
	}

	uid := getUid(c)

	if tf.EntityType == model.EntityComment {
		return singleton.UserLikeService.CommentUnLike(uid, tf.EntityId)
	} else if tf.EntityType == model.EntityTopic {
		return singleton.UserLikeService.TopicUnLike(uid, tf.EntityId)
	} else {
		return 0, singleton.Localizer.ErrorT("entity unsupported")
	}
}

// Post is liked
// @Summary Post is liked
// @Security BearerAuth
// @Schemes
// @Description Post is liked
// @Tags auth required
// @Accept json
// @param request body model.UserLikeForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[bool]
// @Router /isLiked [post]
func isLiked(c *gin.Context) (bool, error) {
	var tf model.UserLikeForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return false, err
	}

	uid := getUid(c)

	return singleton.UserLikeService.Exists(uid, tf.EntityType, tf.EntityId), nil
}

// Post liked ids
// @Summary Post liked ids
// @Security BearerAuth
// @Schemes
// @Description Post liked ids
// @Tags auth required
// @Accept json
// @param request body model.GetLikeIdsForm true "Like Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[[]uint64]
// @Router /likedIds [post]
func likedIds(c *gin.Context) ([]uint64, error) {
	var tf model.GetLikeIdsForm
	if err := c.ShouldBindJSON(&tf); err != nil {
		return nil, err
	}

	uid := getUid(c)

	return singleton.UserLikeService.IsLiked(uid, tf.EntityType, tf.EntityIds)
}
