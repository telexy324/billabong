package model

// 用户点赞
type UserLikeForm struct {
	EntityId   uint64 `json:"entityId"`   // 实体编号
	EntityType int    `json:"entityType"` // 实体类型 	// 创建时间
}

type GetLikeIdsForm struct {
	EntityIds  []uint64 `json:"entityIds"`  // 实体编号
	EntityType int      `json:"entityType"` // 实体类型 	// 创建时间
}
