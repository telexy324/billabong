package model

type Upload struct {
	Common

	Name string `json:"name"` // 文件名
	Url  string `json:"url"`  // 文件地址
	Tag  string `json:"tag"`  // 文件标签
	Key  string `json:"key"`  // 编号
}
