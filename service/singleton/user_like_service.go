package singleton

import (
	"errors"
	"github.com/telexy324/billabong/model"
	"gorm.io/gorm"
)

var UserLikeService = newUserLikeService()

func newUserLikeService() *userLikeService {
	return &userLikeService{}
}

type userLikeService struct {
}

// 统计数量
func (s *userLikeService) Count(entityType string, entityId int64) (count int64, err error) {
	err = DB.Model(&model.UserLike{}).Where("entity_id = ?", entityId).Where("entity_type = ?", entityType).Count(&count).Error
	return
}

// 最近点赞
func (s *userLikeService) Recent(entityType string, entityId int64, count int) (userLikes []model.UserLike, err error) {
	err = DB.Where("entity_id = ?", entityId).Where("entity_type = ?", entityType).Order("id desc").Limit(count).Find(&userLikes).Error
	return
}

// Exists 是否点赞
func (s *userLikeService) Exists(userId uint64, entityType int, entityId uint64) bool {
	if err := DB.Where("user_id = ?", userId).Where("entity_id = ?", entityId).Where("entity_type = ?", entityType).Find(&model.UserLike{}).Error; err != nil {
		return false
	}
	return true
}

// 是否点赞，返回已点赞实体编号
func (s *userLikeService) IsLiked(userId uint64, entityType int, entityIds []uint64) (likedEntityIds []uint64, err error) {
	var userLikes []model.UserLike
	if err = DB.Where("user_id = ?", userId).Where("entity_id in ?", entityIds).Where("entity_type = ?", entityType).Find(&userLikes).Error; err != nil {
		return nil, err
	}
	for _, like := range userLikes {
		likedEntityIds = append(likedEntityIds, like.EntityId)
	}
	return
}

// TopicLike 话题点赞
func (s *userLikeService) TopicLike(userId uint64, topicId uint64) (int64, error) {
	var topic model.Topic
	if err := DB.Where("id = ?", topicId).Find(&topic).Error; err != nil {
		return 0, err
	}
	if topic.Status != model.StatusOk {
		return 0, errors.New("话题不存在")
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := s.like(tx, userId, model.EntityTopic, topicId); err != nil {
			return err
		}
		// 更新点赞数
		var oldTopic model.Topic
		if err := tx.Where("id = ?", topicId).Find(&oldTopic).Error; err != nil {
			return err
		}
		return tx.Update("like_count", oldTopic.LikeCount+1).Error
	}); err != nil {
		return 0, err
	}

	//// 发送事件
	//event.Send(event.UserLikeEvent{
	//	UserId:     userId,
	//	EntityId:   topicId,
	//	EntityType: constants.EntityTopic,
	//})

	return topic.LikeCount + 1, nil
}

func (s *userLikeService) TopicUnLike(userId uint64, topicId uint64) (int64, error) {
	var topic model.Topic
	if err := DB.Where("id = ?", topicId).Find(&topic).Error; err != nil {
		return 0, err
	}
	if topic.Status != model.StatusOk {
		return 0, errors.New("话题不存在")
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := s.like(tx, userId, model.EntityTopic, topicId); err != nil {
			return err
		}
		// 更新点赞数
		var oldTopic model.Topic
		if err := tx.Where("id = ?", topicId).Find(&oldTopic).Error; err != nil {
			return err
		}
		return tx.Update("like_count", oldTopic.LikeCount-1).Error
	}); err != nil {
		return 0, err
	}

	//// 发送事件
	//event.Send(event.UserUnLikeEvent{
	//	UserId:     userId,
	//	EntityId:   topicId,
	//	EntityType: constants.EntityTopic,
	//})

	return topic.LikeCount - 1, nil
}

// CommentLike 话题点赞
func (s *userLikeService) CommentLike(userId uint64, commentId uint64) (int64, error) {
	var comment model.Comment
	if err := DB.Where("id = ?", commentId).Find(&comment).Error; err != nil {
		return 0, err
	}
	if comment.Status != model.StatusOk {
		return 0, errors.New("评论不存在")
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := s.like(tx, userId, model.EntityComment, commentId); err != nil {
			return err
		}
		// 更新点赞数
		var oldComment model.Comment
		if err := tx.Where("id = ?", commentId).Find(&oldComment).Error; err != nil {
			return err
		}
		return tx.Update("like_count", oldComment.LikeCount+1).Error
	}); err != nil {
		return 0, err
	}

	//// 发送事件
	//event.Send(event.UserLikeEvent{
	//	UserId:     userId,
	//	EntityId:   commentId,
	//	EntityType: constants.EntityComment,
	//})

	return comment.LikeCount + 1, nil
}

// CommentLike 话题点赞
func (s *userLikeService) CommentUnLike(userId uint64, commentId uint64) (int64, error) {
	var comment model.Comment
	if err := DB.Where("id = ?", commentId).Find(&comment).Error; err != nil {
		return 0, err
	}
	if comment.Status != model.StatusOk {
		return 0, errors.New("评论不存在")
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := s.like(tx, userId, model.EntityComment, commentId); err != nil {
			return err
		}
		// 更新点赞数
		var oldComment model.Comment
		if err := tx.Where("id = ?", commentId).Find(&oldComment).Error; err != nil {
			return err
		}
		return tx.Update("like_count", oldComment.LikeCount-1).Error
	}); err != nil {
		return 0, err
	}

	//// 发送事件
	//event.Send(event.UserUnLikeEvent{
	//	UserId:     userId,
	//	EntityId:   commentId,
	//	EntityType: constants.EntityComment,
	//})

	return comment.LikeCount - 1, nil
}

func (s *userLikeService) like(tx *gorm.DB, userId uint64, entityType int, entityId uint64) error {
	// 判断是否已经点赞了
	if s.Exists(userId, entityType, entityId) {
		return errors.New("已点赞")
	}
	// 点赞
	var userLike model.UserLike
	userLike.UserID = userId
	userLike.EntityId = entityId
	userLike.EntityType = entityType
	return tx.Create(&userLike).Error
}

func (s *userLikeService) unlike(tx *gorm.DB, userId int64, entityType string, entityId int64) error {
	return tx.Delete(&model.UserLike{}, "user_id = ? and entity_id = ? and entity_type = ?", userId, entityId, entityType).Error
}
