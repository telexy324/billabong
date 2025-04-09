package model

type CommentForm struct {
	EntityType   int      `json:"entityType,omitempty" validate:"optional"`   // 被评论实体类型
	EntityId     int64    `json:"entityId,omitempty" validate:"optional"`     // 被评论实体编号
	Content      string   `json:"content,omitempty" validate:"optional"`      // 内容
	ImageList    string   `json:"imageList,omitempty" validate:"optional"`    // 图片
	ContentType  string   `json:"contentType,omitempty" validate:"optional"`  // 内容类型：markdown、html
	QuoteId      int64    `json:"quoteId,omitempty" validate:"optional"`      // 引用的评论编号
	LikeCount    int64    `json:"likeCount,omitempty" validate:"optional"`    // 点赞数量
	CommentCount int64    `json:"commentCount,omitempty" validate:"optional"` // 评论数量
	Status       int      `json:"status,omitempty" validate:"optional"`       // 状态：0：待审核、1：审核通过、2：审核失败、3：已发布
	Images       []Upload `gorm:"-" json:"images,omitempty" validate:"optional"`
}
