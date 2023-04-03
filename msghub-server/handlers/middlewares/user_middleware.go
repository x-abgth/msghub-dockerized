package middlewares

import (
	"context"
	"log"
	"net/http"

	jwtPkg "github.com/x-abgth/msghub-dockerized/msghub-server/utils/jwt"
)

func UserAuthorizationBeforeLogin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				handler.ServeHTTP(w, r)
			}
		}()

		c, err1 := r.Cookie("user_token")
		if err1 != nil {
			panic("Cookie not found!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if claim.IsAuthenticated {
			log.Println("Redirecting to dashboard!")

			userID := claim.UserID
			ctx := context.WithValue(r.Context(), "userId", userID)
			http.Redirect(w, r.WithContext(ctx), "/user/dashboard", http.StatusFound)
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

		c, err1 := r.Cookie("user_token")
		if err1 != nil {
			panic("Cookie not found!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if !claim.IsAuthenticated {
			panic("redirecting to login page")
		} else {
			userID := claim.UserID
			ctx := context.WithValue(r.Context(), "userId", userID)
			handler.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
