package singleton

import (
	"errors"
	"github.com/telexy324/billabong/model"
	"gorm.io/gorm"
)

var FavoriteService = newFavoriteService()

func newFavoriteService() *favoriteService {
	return &favoriteService{}
}

type favoriteService struct {
}

func (s *favoriteService) Delete(userId uint64, id uint64) error {
	return DB.Model(&model.Favorite{}).Where("user_id = ? and entity_id = ?", userId, id).Delete(&model.Favorite{}).Error
}

func (s *favoriteService) IsFavorited(userId uint64, entityType string, entityId uint64) bool {
	if err := DB.Where("user_id = ?", userId).Where("entity_id = ?", entityId).Where("entity_type = ?", entityType).Find(&model.Favorite{}).Error; err != nil {
		return false
	}
	return true
}

func (s *favoriteService) GetBy(userId uint64, entityType int, entityId uint64) (model.Favorite, error) {
	var favorite model.Favorite
	err := DB.Where("user_id = ? and entity_type = ? and entity_id = ?", userId, entityType, entityId).Find(&favorite).Error
	return favorite, err
}

// AddTopicFavorite 收藏主题
func (s *favoriteService) AddTopicFavorite(userId, topicId uint64) error {
	var topic model.Topic
	if err := DB.Where("id = ?", topicId).Find(&topic).Error; err != nil {
		return err
	}
	if topic.Status != model.StatusOk {
		return errors.New("话题不存在")
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		return s.like(tx, userId, model.EntityTopic, topicId)
	}); err != nil {
		return err
	}
	return nil
}

func (s *favoriteService) like(tx *gorm.DB, userId uint64, entityType int, entityId uint64) error {
	// 判断是否已经点赞了
	isLiked, err := s.Exists(userId, entityType, entityId)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已点赞")
	}
	// 点赞
	var userLike model.Favorite
	userLike.UserID = userId
	userLike.EntityId = entityId
	userLike.EntityType = entityType
	return tx.Create(&userLike).Error
}

func (s *favoriteService) Exists(userId uint64, entityType int, entityId uint64) (bool, error) {
	var count int64
	if err := DB.Where("user_id = ?", userId).Where("entity_id = ?", entityId).Where("entity_type = ?", entityType).Find(&model.Favorite{}).Count(&count).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
