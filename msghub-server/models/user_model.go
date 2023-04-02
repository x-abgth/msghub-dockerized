package models

type UserModel struct {
	UserAvatarUrl string
	UserAbout     string
	UserName      string
	UserPhone     string
	UserPass      string
	UserBlocked   bool
	BlockDur      string
	IsBlocked     bool
	DeletedTime   string
}

type User struct {
	UserPhNo        string  `gorm:"not null;primaryKey;autoIncrement:false" json:"user_ph_no"`
	UserName        string  `gorm:"not null" json:"user_name"`
	UserAvatar      *string `json:"user_avatar"`
	UserAbout       string  `gorm:"not null" json:"user_about"`
	UserPassword    string  `gorm:"not null" json:"user_password"`
	IsBlocked       bool    `gorm:"not null" json:"is_blocked"`
	BlockedDuration *string `json:"blocked_duration"`
	BlockList       *string `json:"block_list"`
}

type DeletedUser struct {
	UserPhNo        string  `gorm:"not null;primaryKey;autoIncrement:false" json:"user_ph_no"`
	UserAvatar      *string `json:"user_avatar"`
	UserAbout       string  `gorm:"not null" json:"user_about"`
	IsBlocked       bool    `gorm:"not null" json:"is_blocked"`
	BlockedDuration *string `json:"blocked_duration"`
	BlockList       *string `json:"block_list"`
	DeleteTime      string  `json:"delete_time"`
}

type Storie struct {
	UserId          string `gorm:"primary key;not null;autoIncrement:false" json:"user_id"`
	StoryUrl        string `gorm:"not null" json:"story_url"`
	StoryUpdateTime string `gorm:"not null" json:"story_update_time"`
	Viewers         string `gorm:"not null" json:"viewers"`
	IsActive        bool   `gorm:"not null" json:"is_active"`
}
