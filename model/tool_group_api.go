package model

type ToolGroupForm struct {
	Name  string   `json:"name" minLength:"1"`
	Tools []uint64 `json:"tools"`
}

type ToolGroupResponseItem struct {
	Group ToolGroup `json:"group"`
	Tools []uint64  `json:"tools"`
}
