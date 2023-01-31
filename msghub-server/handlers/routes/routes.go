package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/x-abgth/msghub-dockerized/msghub-server/socket"
	"github.com/x-abgth/msghub-dockerized/msghub-server/template"
)

func InitializeRoutes(theMux *mux.Router, server *socket.WsServer) {
	userRoutes(theMux, server)
	adminRoutes(theMux)
	theMux.NotFoundHandler = http.HandlerFunc(noPageHandler)
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
