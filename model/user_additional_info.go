package model

type UserAdditionalInfo struct {
	Common
	Avatar       string `gorm:"type:text" json:"avatar,omitempty"`
	Description  string `gorm:"type:text" json:"description,omitempty"` // 个人描述
	Status       int    `json:"status"`                                 // 状态
	TopicCount   int    `json:"topicCount"`                             // 帖子数量
	CommentCount int    `json:"commentCount"`                           // 跟帖数量
	FollowCount  int    `json:"followCount"`                            // 关注数量
	FansCount    int    `json:"fansCount"`                              // 粉丝数量
}
