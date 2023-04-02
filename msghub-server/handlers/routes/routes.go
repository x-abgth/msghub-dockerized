package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/x-abgth/msghub-dockerized/msghub-server/handlers"
	"github.com/x-abgth/msghub-dockerized/msghub-server/logic"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
	"github.com/x-abgth/msghub-dockerized/msghub-server/socket"
	"github.com/x-abgth/msghub-dockerized/msghub-server/template"
	jwtPkg "github.com/x-abgth/msghub-dockerized/msghub-server/utils/jwt"
)

func InitializeRoutes(db *sql.DB, theMux *mux.Router, server *socket.WsServer) {

	userRepository := repository.NewUserRepository(db)
	groupRepository := repository.NewGroupRepository(db)
	migrationRepository := repository.NewMigrationRepository(db)
	messageRepository := repository.NewMessageRepository(db)
	adminRepository := repository.NewAdminRepository(db)

	migrationService := logic.NewMigrateLogic(migrationRepository)
	userService := logic.NewUserLogic(userRepository, groupRepository, messageRepository)
	adminService := logic.NewAdminLogic(adminRepository, userRepository, messageRepository)

	userHandler := handlers.NewUserHandler(migrationService, userService)
	adminHandler := handlers.NewAdminHandler(migrationService, adminService)

	socket.NewSocketRepositoryMethods(userService)

	userRoutes(theMux, userHandler, server)
	adminRoutes(theMux, adminHandler)
	theMux.NotFoundHandler = http.HandlerFunc(noPageHandler)

	// WEBSOCKET CONNECTIONS
	hub := &socket.Hub{
		Clients:    make(map[string]map[*socket.GClient]bool),
		Register:   make(chan *socket.GClient),
		Unregister: make(chan *socket.GClient),
		Broadcast:  make(chan *socket.WSMessage),
	}
	go hub.Run()

	// For personal messaging
	theMux.HandleFunc("/ws/{target}", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				log.Println(e)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}()

		vars := mux.Vars(r)
		target := vars["target"]

		c, err1 := r.Cookie("user_token")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("No cookie found - " + err1.Error())
			}
			panic(err1.Error())
		}

		claim := jwtPkg.GetValueFromJwt(c) // error
		if claim == nil {
			panic("JWT error happened!")
		}

		socket.ServeWs(claim.User.UserPhone, target, server, w, r)
	})

	// For group messaging
	theMux.HandleFunc("/ws/group/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--------- IN /WS/TARGET HANDLER FUNCTION ------------")

		defer func() {
			if e := recover(); e != nil {
				log.Println(e)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}()

		vars := mux.Vars(r)
		target := vars["id"]

		c, err1 := r.Cookie("user_token")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("No cookie found - " + err1.Error())
			}
			panic(err1.Error())
		}

		claim := jwtPkg.GetValueFromJwt(c) // error

		if !userService.CheckUserLeftTheGroup(claim.User.UserPhone, target) {
			socket.ServeGroupWs(hub, target, w, r)
		}
	})
}

// 404 Error page handler function
func noPageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "404 Error Page",
	}

	err := template.Tpl.ExecuteTemplate(w, "error_page.html", data)
	if err != nil {
		log.Fatal("Couldn't render the error page handler!")
	}
}
