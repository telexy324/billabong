package model

type Favorite struct {
	Common

	EntityType int   `json:"entityType"` // 收藏实体类型
	EntityId   int64 `json:"entityId"`   // 收藏实体编号
}
