package model

import (
	"encoding/json"
	"gorm.io/gorm"
)

// EntityType
const (
	EntityTopic = iota + 1
	EntityComment
)

type Comment struct {
	Common

	EntityType   int      `json:"entityType"`   // 被评论实体类型
	EntityId     int64    `json:"entityId"`     // 被评论实体编号
	Content      string   `json:"content"`      // 内容
	ImageList    string   `json:"imageList"`    // 图片
	ContentType  string   `json:"contentType"`  // 内容类型：markdown、html
	QuoteId      int64    `json:"quoteId"`      // 引用的评论编号
	LikeCount    int64    `json:"likeCount"`    // 点赞数量
	CommentCount int64    `json:"commentCount"` // 评论数量
	Status       int      `json:"status"`       // 状态：0：待审核、1：审核通过、2：审核失败、3：已发布
	Images       []Upload `gorm:"-" json:"images"`
	Liked        bool     `gorm:"-" json:"liked"`
	Favorited    bool     `gorm:"-" json:"favorited"`
}

func (m *Comment) BeforeSave(tx *gorm.DB) error {
	if m.Images != nil && len(m.Images) > 0 {
		fileIds := make([]uint64, 0, len(m.Images))
		for _, file := range m.Images {
			fileIds = append(fileIds, file.ID)
		}
	}
	if data, err := json.Marshal(m.Images); err != nil {
		return err
	} else {
		m.ImageList = string(data)
	}
	return nil
}

func (m *Comment) AfterFind(tx *gorm.DB) error {
	fileIds := make([]uint64, 0)
	if err := json.Unmarshal([]byte(m.ImageList), &fileIds); err != nil {
		return err
	}
	if err := tx.Model(&Upload{}).Where("id in (?)", fileIds).Find(&m.Images).Error; err != nil {
		return err
	}
	return nil
}
