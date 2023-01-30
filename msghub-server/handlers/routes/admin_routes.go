package routes

import (
	"github.com/x-abgth/msghub/msghub-server/handlers"
	"github.com/x-abgth/msghub/msghub-server/handlers/middlewares"

	"github.com/gorilla/mux"
)

func adminRoutes(theMux *mux.Router) {
	adminHandlerInfo := handlers.AdminHandlerStruct{}

	// OTHER HANDLERS.
	admin := theMux.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login-page", adminHandlerInfo.AdminLoginPageHandler).Methods("GET")
	admin.HandleFunc("/login-page", adminHandlerInfo.AdminAuthenticateHandler).Methods("POST")
	admin.HandleFunc("/dashboard", middlewares.AdminAuthenticationMiddleware(adminHandlerInfo.AdminDashboardHandler)).Methods("GET")
	admin.HandleFunc("/user-block/{id}/{condition}", adminHandlerInfo.AdminBlocksUserHandler).Methods("GET")
	admin.HandleFunc("/user-unblock/{id}", adminHandlerInfo.AdminUnBlocksUserHandler).Methods("GET")
	admin.HandleFunc("/group-block/{id}/{condition}", adminHandlerInfo.AdminBlocksGroupHandler).Methods("GET")
	admin.HandleFunc("/group-unblock/{id}", adminHandlerInfo.AdminUnBlockGroupHandler).Methods("GET")
	admin.HandleFunc("/broadcast-message", adminHandlerInfo.AdminBroadcastHandler).Methods("POST")
	admin.HandleFunc("/new-admin", adminHandlerInfo.NewAdminPageHandler).Methods("GET")
	admin.HandleFunc("/new-admin", adminHandlerInfo.NewAdminHandler).Methods("POST")
	admin.HandleFunc("/logout", adminHandlerInfo.AdminLogoutHandler).Methods("GET")
}
