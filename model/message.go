package model

// 消息状态
const (
	StatusUnread   = iota // 消息未读
	StatusHaveRead        // 消息已读
)

const (
	TypeTopicComment   = iota // 收到话题评论
	TypeCommentReply          // 收到他人回复
	TypeTopicLike             // 收到点赞
	TypeTopicFavorite         // 话题被收藏
	TypeTopicRecommend        // 话题被设为推荐
	TypeTopicDelete           // 话题被删除
	TypeArticleComment        // 收到文章评论
)

type Message struct {
	Common

	FromId       int64  `json:"fromId"`       // 消息发送人
	UserId       int64  `json:"userId"`       // 用户编号(消息接收人)
	Title        string `json:"title"`        // 消息标题
	Content      string `json:"content"`      // 消息内容
	QuoteContent string `json:"quoteContent"` // 引用内容
	Type         int    `json:"type"`         // 消息类型
	Status       int    `json:"status"`       // 状态：0：未读、1：已读
}
