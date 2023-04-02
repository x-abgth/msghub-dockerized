package routes

import (
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers"
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers/middlewares"
	"github.com/x-abgth/msghub-dockerized/msghub-server/socket"

	"github.com/gorilla/mux"
)

func userRoutes(theMux *mux.Router, userHandler *handlers.UserHandler, s *socket.WsServer) {

	theMux.HandleFunc("/register/phone", userHandler.UserVerifyRegisterPhoneHandler).Methods("POST")
	theMux.HandleFunc("/register", userHandler.UserVerifyRegisterOtpHandler).Methods("POST")

	// login and register functions
	theMux.HandleFunc("/", middlewares.UserAuthorizationBeforeLogin(userHandler.UserLoginHandler)).Methods("GET")
	theMux.HandleFunc("/", userHandler.UserLoginCredentialsHandler).Methods("POST")
	theMux.HandleFunc("/login/phone/validation", userHandler.UserLoginWithOtpPageHandler).Methods("GET")
	theMux.HandleFunc("/login/phone/validation", userHandler.UserVerifyLoginPhoneHandler).Methods("POST")
	theMux.HandleFunc("/login/otp/validation", userHandler.UserVerifyLoginOtpHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard", middlewares.UserAuthorizationAfterLogin(userHandler.UserDashboardHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/people", middlewares.UserAuthorizationAfterLogin(userHandler.UserShowPeopleHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/new-chat-started/{target}", middlewares.UserAuthorizationAfterLogin(userHandler.UserNewChatStartedHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", userHandler.UserProfilePageHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", userHandler.UserProfileUpdateHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/add-story/{target}", userHandler.UserAddStoryHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/delete-story/{target}", userHandler.UserDeleteStoryHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/story-seen/{target}", userHandler.UserStorySeenHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/create-group", userHandler.UserCreateGroup).Methods("POST")
	theMux.HandleFunc("/user/dashboard/add-group-members", userHandler.UserAddGroupMembers).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-created-finally", userHandler.UserGroupCreationHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/chat-selected", userHandler.UserNewChatSelectedHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-chat-selected", userHandler.UserGroupChatSelectedHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-manage-members/{target}", userHandler.UserGroupManagePageHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-members-added/{target}", userHandler.UserGroupAddMembersHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-got-unblocked", userHandler.GroupUnblockHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-left/{target}", userHandler.UserLeftGroupHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-kicked-out/{group}/{user}", userHandler.UserKickedOutHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-block-user/{target}", userHandler.UserBlocksHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-unblock-user/{target}", userHandler.UserUnblocksHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/about-page", userHandler.AboutPageHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/delete-account/{target}", userHandler.UserDeleteAccountHandler).Methods("GET")

	theMux.HandleFunc("/user/logout", userHandler.UserLogoutHandler).Methods("GET")
}
