package middlewares

import (
	"context"
	"net/http"

	jwtPkg "github.com/x-abgth/msghub-dockerized/msghub-server/utils/jwt"
)

func AdminAuthenticationMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				http.Redirect(w, r, "/admin/login-page", http.StatusFound)
			}
		}()

		c, err1 := r.Cookie("admin_token")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("Cookie not found!")
			}
			panic("Unknown error occurred!")
		}

		claim := jwtPkg.GetValueFromAdminJwt(c)
		if !claim.IsAuthenticated {
			panic("redirecting to login page")
		} else {
			admin := claim.AdminName
			ctx := context.WithValue(r.Context(), "admin", admin)
			handler.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
