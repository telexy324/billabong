package model

import (
	"encoding/json"
	"gorm.io/gorm"
)

type Topic struct {
	Common
	// 用户
	Title             string   `json:"title" form:"title"`                                           // 标题
	Content           string   `gorm:"type:longtext" json:"content" form:"content"`                  // 内容
	ImageList         string   `gorm:"type:longtext default:'[]'" json:"imageList" form:"imageList"` // 图片 	// 回复可见内容
	Recommend         bool     `json:"recommend" form:"recommend"`                                   // 是否推荐
	RecommendTime     int64    `json:"recommendTime" form:"recommendTime"`                           // 推荐时间
	Sticky            bool     `json:"sticky" form:"sticky"`                                         // 置顶
	StickyTime        int64    `json:"stickyTime" form:"stickyTime"`                                 // 置顶时间
	ViewCount         int64    `json:"viewCount" form:"viewCount"`                                   // 查看数量
	CommentCount      int64    `json:"commentCount" form:"commentCount"`                             // 跟帖数量
	LikeCount         int64    `json:"likeCount" form:"likeCount"`                                   // 点赞数量
	Status            int      `json:"status" form:"status"`                                         // 状态：0：正常、1：删除
	LastCommentTime   int64    `json:"lastCommentTime" form:"lastCommentTime"`                       // 最后回复时间
	LastCommentUserId int64    `json:"lastCommentUserId" form:"lastCommentUserId"`                   // 最后回复用户 	// 扩展数据
	Images            []Upload `gorm:"-" json:"images"`
}

func (m *Topic) BeforeSave(tx *gorm.DB) error {
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

func (m *Topic) AfterFind(tx *gorm.DB) error {
	fileIds := make([]uint64, 0)
	if err := json.Unmarshal([]byte(m.ImageList), &fileIds); err != nil {
		return err
	}
	if err := tx.Model(&Upload{}).Where("id in (?)", fileIds).Find(&m.Images).Error; err != nil {
		return err
	}
	return nil
}
