package models

type GroupModel struct {
	Id          string
	Owner       string
	Image       string
	Name        string
	About       string
	CreatedDate string
	NoOfMembers int
	IsBanned    bool
	BanTime     string
	Members     []string
}

type GroupMessageModel struct {
	MsgId    string
	GroupId  string
	SenderId string
	Content  string
	Type     string
	Status   string
	Time     string
}

type GroupMembersModel struct {
	MPhone   string `json:"phone"`
	MName    string `json:"name"`
	MAvatar  string `json:"avatar"`
	MIsAdmin bool   `json:"is_admin"`
}

type ManageGroupMember struct {
	UserAvatarUrl string `json:"avatar"`
	UserName      string `json:"name"`
	UserPhone     string `json:"phone"`
	UserAbout     string `json:"about"`
	IsMember      bool   `json:"is_member"`
}

type Group struct {
	GroupId           int    `json:"group_id"`
	GroupName         string `json:"group_name"`
	GroupAvatar       string `json:"group_avatar"`
	GroupAbout        string `json:"group_about"`
	GroupCreator      string `json:"group_creator"`
	GroupCreatedDate  string `json:"group_created_date"`
	GroupTotalMembers int    `json:"group_total_members"`
	IsBanned          bool   `json:"is_banned"`
	BannedTime        string `json:"banned_time"`
}

type UserGroupRelation struct {
	Id       int    `json:"id"`
	GroupId  int    `json:"group_id"`
	UserId   string `json:"user_id"`
	UserRole string `json:"user_role"`
}

type GroupMessage struct {
	MsgId          int    `json:"msg_id"`
	GroupId        int    `json:"group_id"`
	SenderId       string `json:"sender_id"`
	MessageContent string `json:"message_content"`
	ContentType    string `json:"content_type"`
	Status         string `json:"status"`
	SentTime       string `json:"sent_time"`
	IsRecent       bool   `json:"is_recent"`
}
