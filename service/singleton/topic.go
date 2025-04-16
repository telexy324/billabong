package singleton

import (
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/pkg/markdown"
)

var TopicService = newTopicService()

func newTopicService() *topicService {
	return &topicService{}
}

type topicService struct {
}

func (s *topicService) BuildTopic(topic model.Topic, uid uint64) (model.Topic, error) {
	topic.Content = markdown.ToHTML(topic.Content)
	favorited, err := FavoriteService.Exists(uid, model.EntityTopic, topic.ID)
	if err != nil {
		return topic, err
	}
	liked, err := UserLikeService.Exists(uid, model.EntityTopic, topic.ID)
	if err != nil {
		return topic, err
	}
	var user model.User
	err = DB.Where("id = ?", topic.UserID).First(&user).Error
	if err != nil {
		return topic, err
	}
	topic.Liked, topic.Favorited, topic.UserName = liked, favorited, user.Username
	return topic, nil
}
