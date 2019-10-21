package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

const (
	secretKey = "This is an admin authorization"
)

// MiddleWare middleware for resource handler
func MiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validateTokenMiddleware(w, r, next.ServeHTTP)
	})
}

func validateTokenMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	if r.UserAgent() != "nginx" {
		next(w, r)
	}
	// nginx 转发需要验证token
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

	if err == nil {
		if token.Valid {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
			log.Print("Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized access to this resource")
	}

}
