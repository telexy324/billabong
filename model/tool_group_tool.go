package model

type ToolGroupTool struct {
	Common
	ToolGroupId uint64 `json:"tool_group_id" gorm:"uniqueIndex:idx_tool_group_tool"`
	ToolId      uint64 `json:"tool_id" gorm:"uniqueIndex:idx_tool_group_tool"`
}
