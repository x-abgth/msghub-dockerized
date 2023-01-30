package middlewares

import (
	"log"
	"net/http"

	jwtPkg "github.com/x-abgth/msghub/msghub-server/utils/jwt"
)

func UserAuthorizationBeforeLogin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				handler.ServeHTTP(w, r)
			}
		}()

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("Cookie not found!")
			}
			panic("Unknown error occurred!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if claim.IsAuthenticated {
			log.Println("Redirecting to dashboard!")
			http.Redirect(w, r, "/user/dashboard", http.StatusFound)
		} else {
			panic("User is not authenticated!")
		}
	}
}

func UserAuthorizationAfterLogin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}()

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("Cookie not found!")
			}
			panic("Unknown error occurred!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if !claim.IsAuthenticated {
			panic("redirecting to login page")
		} else {
			handler.ServeHTTP(w, r)
		}
	}
}
