package model

type TopicGroupTopic struct {
	Common
	TopicGroupId uint64 `json:"topic_group_id" gorm:"uniqueIndex:idx_topic_group_topic"`
	TopicId      uint64 `json:"topic_id" gorm:"uniqueIndex:idx_topic_group_topic"`
}
