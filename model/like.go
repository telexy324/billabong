package model

// 用户点赞
type UserLike struct {
	Common
	EntityId   int64 `json:"topicId"`    // 实体编号
	EntityType int   `json:"entityType"` // 实体类型 	// 创建时间
}
