package model

type TopicGroupForm struct {
	Name   string   `json:"name" minLength:"1"`
	Topics []uint64 `json:"topics"`
}

type TopicGroupResponseItem struct {
	Group  TopicGroup `json:"group"`
	Topics []uint64   `json:"topics"`
}
