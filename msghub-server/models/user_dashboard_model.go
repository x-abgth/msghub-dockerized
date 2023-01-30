package models

type RecentMessages struct {
	Id          string
	Name        string
	Avatar      string
	LastMsg     string
	LastMsgTime string
	IsImage     bool
	IsRead      bool
}

type RecentChatModel struct {
	Content   RecentMessages
	Sender    string
	IsGroup   bool
	Order     float64
	IsBlocked bool
	IsOnline  bool
}

type StoryModel struct {
	UserName    string
	UserPhone   string
	UserAvatar  string
	StoryImg    string
	ViewerCount int
	Viewers     []UserModel
	Expiration  string
	IsViewed    bool
}

type UserDashboardModel struct {
	UserPhone      string
	UserDetails    UserModel
	UserStory      StoryModel
	RecentChatList []RecentChatModel
	StoryList      []StoryModel
}
