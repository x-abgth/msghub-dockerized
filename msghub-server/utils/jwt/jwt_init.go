package jwt

import (
	"net/http"
	"os"

	"github.com/x-abgth/msghub/msghub-server/models"

	"github.com/golang-jwt/jwt/v4"
)

type UserJwtClaim struct {
	User            models.UserModel
	GroupModel      models.GroupModel
	IsAuthenticated bool
	jwt.RegisteredClaims
}

type AdminJwtClaim struct {
	AdminName       string
	IsAuthenticated bool
	jwt.RegisteredClaims
}

var JwtKey []byte

func InitJwtKey() {

	key := os.Getenv("JWT_KEY")

	JwtKey = []byte(key)
}

// UserJwtToken will
func SignJwtToken(u *UserJwtClaim) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		panic("Internal server error!")
	}
	return tokenString
}

func GetValueFromJwt(c *http.Cookie) *UserJwtClaim {
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &UserJwtClaim{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil {
		return nil
	}
	if !tkn.Valid {
		return nil
	}
	return claims
}

func SignAdminJwtToken(u *AdminJwtClaim) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		panic("Internal server error!")
	}
	return tokenString
}

func GetValueFromAdminJwt(c *http.Cookie) *AdminJwtClaim {
	tknStr := c.Value

	claims := &AdminJwtClaim{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil {
		return nil
	}
	if !tkn.Valid {
		return nil
	}
	return claims
}
