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

var userVal *UserModel

func InitUserModel(model UserModel) *UserModel {
	userVal = &UserModel{
		UserAvatarUrl: model.UserAvatarUrl,
		UserAbout:     model.UserAbout,
		UserName:      model.UserName,
		UserPhone:     model.UserPhone,
		UserPass:      model.UserPass,
	}
	return userVal
}

func ReturnUserModel() *UserModel {
	return userVal
}
