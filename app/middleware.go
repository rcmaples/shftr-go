package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"shftr/helpers"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}

func checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := contextKey("user")
		var tokenString string

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Split(authHeader, "Bearer ")[1]
		} else {
			cookieToken, err := r.Cookie("token")
			if err != http.ErrNoCookie && err != nil {
				helpers.Logger.Println("cookie error: ", err)
				errorJson(w, err, http.StatusUnauthorized)
				return
			} else if err == http.ErrNoCookie {
				helpers.Logger.Println("no auth cookie found")
				err := errors.New("no auth token")
				errorJson(w, err, http.StatusUnauthorized)
				return
			}

			tokenString = cookieToken.Value
		}

		if tokenString == "" {
			helpers.Logger.Println("no token found in cookie or auth header")
			err := errors.New("no auth token")
			errorJson(w, err, http.StatusUnauthorized)
			return
		}

		token, err := decodeToken(tokenString)
		if err != nil {
			helpers.Logger.Println("error decoding token: ", err)
			errorJson(w, err, http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), u, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		} else {
			helpers.Logger.Println("error processing token claims: ", err)
			errorJson(w, err, http.StatusUnauthorized)
			return
		}
	})
}

func decodeToken(ts string) (*jwt.Token, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(ts, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
