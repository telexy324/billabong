package model

type UserFavoriteForm struct {
	EntityId   uint64 `json:"entityId"`   // 实体编号
	EntityType int    `json:"entityType"` // 实体类型 	// 创建时间
}
