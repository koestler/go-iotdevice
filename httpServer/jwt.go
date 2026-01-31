package httpServer

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type jwtClaims struct {
	User string `json:"sub"`
	jwt.RegisteredClaims
}

func createJwtToken(config AuthenticationConfig, user string) (tokenStr string, err error) {
	claims := &jwtClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JwtValidityPeriod())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(config.JwtSecret())
}

func authJwtMiddleware(env *Environment) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract jwt token from authorization header if present
			tokenStr := r.Header.Get("Authorization")
			if len(tokenStr) < 1 {
				next.ServeHTTP(w, r)
				return
			}

			if user, err := checkToken(tokenStr, env.Authentication.JwtSecret()); err != nil {
				jsonErrorResponse(w, http.StatusUnauthorized, errors.New("invalid jwt token"))
				return
			} else {
				// continue; if user is set this means a valid token is present
				ctx := context.WithValue(r.Context(), authUserKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

type contextKey string

const authUserKey contextKey = "AuthUser"

func checkToken(tokenStr string, jwtSecret []byte) (user string, err error) {
	// decode jwt token
	claims := &jwtClaims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	if !tkn.Valid {
		return "", errors.New("invalid token")
	}
	return claims.User, nil
}

func isViewAuthenticated(view ViewConfig, r *http.Request, allowAnonymous bool) bool {
	user, _ := r.Context().Value(authUserKey).(string)
	return isViewAuthenticatedByUser(view, user, allowAnonymous)
}

func isViewAuthenticatedByUser(view ViewConfig, user string, allowAnonymous bool) bool {
	if !allowAnonymous && len(user) < 1 {
		return false
	}

	if view.IsPublic() {
		return true
	}
	if len(user) < 1 {
		return false
	}

	return view.IsAllowed(user)
}
