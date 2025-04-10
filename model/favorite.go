package model

type Favorite struct {
	Common

	EntityType uint  `json:"entityType"` // 收藏实体类型
	EntityId   int64 `json:"entityId"`   // 收藏实体编号
}
