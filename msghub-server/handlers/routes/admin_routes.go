package routes

import (
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers"
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers/middlewares"

	"github.com/gorilla/mux"
)

func adminRoutes(theMux *mux.Router, adminHandler *handlers.AdminHandler) {

	// OTHER HANDLERS.
	admin := theMux.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login-page", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminLoginPageHandler)).Methods("GET")
	admin.HandleFunc("/login-page", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminAuthenticateHandler)).Methods("POST")
	admin.HandleFunc("/dashboard", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminDashboardHandler)).Methods("GET")
	admin.HandleFunc("/user-block/{id}/{condition}", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminBlocksUserHandler)).Methods("GET")
	admin.HandleFunc("/user-unblock/{id}", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminUnBlocksUserHandler)).Methods("GET")
	admin.HandleFunc("/group-block/{id}/{condition}", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminBlocksGroupHandler)).Methods("GET")
	admin.HandleFunc("/group-unblock/{id}", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminUnBlockGroupHandler)).Methods("GET")
	admin.HandleFunc("/broadcast-message", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminBroadcastHandler)).Methods("POST")
	admin.HandleFunc("/new-admin", middlewares.AdminAuthenticationMiddleware(adminHandler.NewAdminPageHandler)).Methods("GET")
	admin.HandleFunc("/new-admin", middlewares.AdminAuthenticationMiddleware(adminHandler.NewAdminHandler)).Methods("POST")
	admin.HandleFunc("/logout", middlewares.AdminAuthenticationMiddleware(adminHandler.AdminLogoutHandler)).Methods("GET")
}
