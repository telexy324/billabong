package model

type TopicForm struct {
	Title             string   `json:"title,omitempty"`                                 // 标题
	Content           string   `json:"content,omitempty"`                               // 内容 	// 图片 	// 回复可见内容
	Recommend         bool     `json:"recommend,omitempty" validate:"optional"`         // 是否推荐
	RecommendTime     int64    `json:"recommendTime,omitempty" validate:"optional"`     // 推荐时间
	Sticky            bool     `json:"sticky,omitempty" validate:"optional"`            // 置顶
	StickyTime        int64    `json:"stickyTime,omitempty" validate:"optional"`        // 置顶时间
	ViewCount         int64    `json:"viewCount,omitempty" validate:"optional"`         // 查看数量
	CommentCount      int64    `json:"commentCount,omitempty" validate:"optional"`      // 跟帖数量
	LikeCount         int64    `json:"likeCount,omitempty" validate:"optional"`         // 点赞数量
	Status            int      `json:"status,omitempty" validate:"optional"`            // 状态：0：正常、1：删除
	LastCommentTime   int64    `json:"lastCommentTime,omitempty" validate:"optional"`   // 最后回复时间
	LastCommentUserId int64    `json:"lastCommentUserId,omitempty" validate:"optional"` // 最后回复用户 	// 扩展数据
	Images            []Upload `json:"images,omitempty" validate:"optional"`
}
