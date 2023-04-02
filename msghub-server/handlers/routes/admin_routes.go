package routes

import (
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers"
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers/middlewares"

	"github.com/gorilla/mux"
)

func adminRoutes(theMux *mux.Router, adminHandler *handlers.AdminHandler) {

	// OTHER HANDLERS.
	admin := theMux.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login-page", adminHandler.AdminLoginPageHandler).Methods("GET")
	admin.HandleFunc("/login-page", adminHandler.AdminAuthenticateHandler).Methods("POST")
	admin.HandleFunc("/dashboard", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminDashboardHandler)).Methods("GET")
	admin.HandleFunc("/user-block/{id}/{condition}", adminHandler.AdminBlocksUserHandler).Methods("GET")
	admin.HandleFunc("/user-unblock/{id}", adminHandler.AdminUnBlocksUserHandler).Methods("GET")
	admin.HandleFunc("/group-block/{id}/{condition}", adminHandler.AdminBlocksGroupHandler).Methods("GET")
	admin.HandleFunc("/group-unblock/{id}", adminHandler.AdminUnBlockGroupHandler).Methods("GET")
	admin.HandleFunc("/broadcast-message", adminHandler.AdminBroadcastHandler).Methods("POST")
	admin.HandleFunc("/new-admin", adminHandler.NewAdminPageHandler).Methods("GET")
	admin.HandleFunc("/new-admin", adminHandler.NewAdminHandler).Methods("POST")
	admin.HandleFunc("/logout", adminHandler.AdminLogoutHandler).Methods("GET")
}
