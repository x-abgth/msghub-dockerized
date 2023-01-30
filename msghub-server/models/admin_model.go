package models

type AdminModel struct {
	AdminId   string
	AdminName string
}

type AdminDashboardModel struct {
	AdminName             string
	UsersTbContent        []UserModel
	DeletedUsersTbContent []UserModel
	AdminTbContent        []AdminModel
	GroupTbContent        []GroupModel
}
