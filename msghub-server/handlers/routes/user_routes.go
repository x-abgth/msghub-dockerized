package routes

import (
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers"
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers/middlewares"
	"github.com/x-abgth/msghub-dockerized/msghub-server/socket"

	"github.com/gorilla/mux"
)

func userRoutes(theMux *mux.Router, userHandler *handlers.UserHandler, s *socket.WsServer) {

	theMux.HandleFunc("/register/phone", middlewares.UserAuthorizationBeforeLogin(userHandler.UserVerifyRegisterPhoneHandler)).Methods("POST")
	theMux.HandleFunc("/register", middlewares.UserAuthorizationBeforeLogin(userHandler.UserVerifyRegisterOtpHandler)).Methods("POST")

	// login and register functions
	theMux.HandleFunc("/", middlewares.UserAuthorizationBeforeLogin(userHandler.UserLoginHandler)).Methods("GET")
	theMux.HandleFunc("/", middlewares.UserAuthorizationBeforeLogin(userHandler.UserLoginCredentialsHandler)).Methods("POST")
	theMux.HandleFunc("/login/phone/validation", middlewares.UserAuthorizationBeforeLogin(userHandler.UserLoginWithOtpPageHandler)).Methods("GET")
	theMux.HandleFunc("/login/phone/validation", middlewares.UserAuthorizationBeforeLogin(userHandler.UserVerifyLoginPhoneHandler)).Methods("POST")
	theMux.HandleFunc("/login/otp/validation", middlewares.UserAuthorizationBeforeLogin(userHandler.UserVerifyLoginOtpHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard", middlewares.UserAuthorizationAfterLogin(userHandler.UserDashboardHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/people", middlewares.UserAuthorizationAfterLogin(userHandler.UserShowPeopleHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/new-chat-started/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserNewChatStartedHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", middlewares.UserAuthorizationAfterLogin(userHandler.UserProfilePageHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", middlewares.UserAuthorizationAfterLogin(userHandler.UserProfileUpdateHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/add-story/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserAddStoryHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/delete-story/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserDeleteStoryHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/story-seen/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserStorySeenHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/create-group", middlewares.UserAuthorizationAfterLogin(userHandler.UserCreateGroup)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/add-group-members", middlewares.UserAuthorizationAfterLogin(userHandler.UserAddGroupMembers)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-created-finally", middlewares.UserAuthorizationAfterLogin(userHandler.UserGroupCreationHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/chat-selected", middlewares.UserAuthorizationAfterLogin(userHandler.UserNewChatSelectedHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-chat-selected", middlewares.UserAuthorizationAfterLogin(userHandler.UserGroupChatSelectedHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-manage-members/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserGroupManagePageHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-members-added/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserGroupAddMembersHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-got-unblocked", middlewares.UserAuthorizationAfterLogin(userHandler.GroupUnblockHandler)).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-left/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserLeftGroupHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-kicked-out/{group}/{user}", middlewares.UserAuthorizationAfterLogin(userHandler.UserKickedOutHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-block-user/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserBlocksHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-unblock-user/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserUnblocksHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/about-page", middlewares.UserAuthorizationAfterLogin(userHandler.AboutPageHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/delete-account/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserDeleteAccountHandler)).Methods("GET")

	theMux.HandleFunc("/user/logout", middlewares.UserAuthorizationAfterLogin(userHandler.UserLogoutHandler)).Methods("GET")
}
