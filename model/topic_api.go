package model

type TopicForm struct {
	Title     string `json:"title,omitempty"`                         // 标题
	Content   string `json:"content,omitempty"`                       // 内容 	// 图片 	// 回复可见内容
	Recommend bool   `json:"recommend,omitempty" validate:"optional"` // 是否推荐
	//RecommendTime     time.Time `json:"recommendTime,omitempty" validate:"optional"`     // 推荐时间
	Sticky bool `json:"sticky,omitempty" validate:"optional"` // 置顶
	//StickyTime        time.Time `json:"stickyTime,omitempty" validate:"optional"`        // 置顶时间
	ViewCount    int64 `json:"viewCount,omitempty" validate:"optional"`    // 查看数量
	CommentCount int64 `json:"commentCount,omitempty" validate:"optional"` // 跟帖数量
	LikeCount    int64 `json:"likeCount,omitempty" validate:"optional"`    // 点赞数量
	Status       int   `json:"status,omitempty" validate:"optional"`       // 状态：0：正常、1：删除
	//LastCommentTime   time.Time `json:"lastCommentTime,omitempty" validate:"optional"`   // 最后回复时间
	LastCommentUserId uint64   `json:"lastCommentUserId,omitempty" validate:"optional"` // 最后回复用户 	// 扩展数据
	Affixes           []Upload `json:"affixes,omitempty" validate:"optional"`
	TopicGroup        uint64   `json:"topicGroup,omitempty" validate:"optional"`
}
