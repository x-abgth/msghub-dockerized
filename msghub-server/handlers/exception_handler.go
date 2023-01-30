package handlers

import (
	"fmt"
	"net/http"
)

func handleExceptions(w http.ResponseWriter, r *http.Request, route string) {
	if e := recover(); e != nil {
		fmt.Println("Recovered from panic : ", e)
		http.Redirect(w, r, route, http.StatusSeeOther)
	}
}
