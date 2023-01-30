package middlewares

import (
	"net/http"

	jwtPkg "github.com/x-abgth/msghub/msghub-server/utils/jwt"
)

func AdminAuthenticationMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				http.Redirect(w, r, "/admin/login-page", http.StatusFound)
			}
		}()

		c, err1 := r.Cookie("adminToken")
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
			handler.ServeHTTP(w, r)
		}
	}
}
