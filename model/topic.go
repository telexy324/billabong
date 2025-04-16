package model

import (
	"database/sql"
	"encoding/json"
	"gorm.io/gorm"
)

const (
	StatusOk      = iota // 正常
	StatusDeleted        // 删除
	StatusReview         // 待审核
)

type Topic struct {
	Common
	// 用户
	Title             string       `json:"title" form:"title"`                              // 标题
	Content           string       `gorm:"type:longtext" json:"content" form:"content"`     // 内容
	AffixList         string       `gorm:"type:longtext" json:"affixList" form:"affixList"` // 图片 	// 回复可见内容
	Recommend         bool         `json:"recommend" form:"recommend"`                      // 是否推荐
	RecommendTime     sql.NullTime `json:"recommendTime" form:"recommendTime"`              // 推荐时间
	Sticky            bool         `json:"sticky" form:"sticky"`                            // 置顶
	StickyTime        sql.NullTime `json:"stickyTime" form:"stickyTime"`                    // 置顶时间
	ViewCount         int64        `json:"viewCount" form:"viewCount"`                      // 查看数量
	CommentCount      int64        `json:"commentCount" form:"commentCount"`                // 跟帖数量
	LikeCount         int64        `json:"likeCount" form:"likeCount"`                      // 点赞数量
	Status            int          `json:"status" form:"status"`                            // 状态：0：正常、1：删除
	LastCommentTime   sql.NullTime `json:"lastCommentTime" form:"lastCommentTime"`          // 最后回复时间
	LastCommentUserId uint64       `json:"lastCommentUserId" form:"lastCommentUserId"`      // 最后回复用户 	// 扩展数据
	Affixes           []Upload     `gorm:"-" json:"affixes"`
	Liked             bool         `gorm:"-" json:"liked"`
	Favorited         bool         `gorm:"-" json:"favorited"`
	UserName          string       `gorm:"-" json:"userName"`
}

func (m *Topic) BeforeSave(tx *gorm.DB) error {
	if m.Affixes != nil && len(m.Affixes) > 0 {
		fileIds := make([]uint64, 0, len(m.Affixes))
		for _, file := range m.Affixes {
			fileIds = append(fileIds, file.ID)
		}
	}
	if data, err := json.Marshal(m.Affixes); err != nil {
		return err
	} else {
		m.AffixList = string(data)
	}
	return nil
}

func (m *Topic) AfterFind(tx *gorm.DB) error {
	if len(m.AffixList) <= 0 {
		return nil
	}
	m.Affixes = make([]Upload, 0)
	return json.Unmarshal([]byte(m.AffixList), &m.Affixes)
}
