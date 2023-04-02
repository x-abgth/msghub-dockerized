package socket

import "github.com/x-abgth/msghub-dockerized/msghub-server/logic"

var userLogic logic.UserLogic

func NewSocketRepositoryMethods(userServ logic.UserLogic) {
	userLogic = userServ
}
