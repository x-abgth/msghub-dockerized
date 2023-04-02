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

type Admin struct {
	AdminId   int    `son:"admin_id"`
	AdminName string `json:"admin_name"`
	AdminPass string `json:"admin_pass"`
}
