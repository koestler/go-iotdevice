package httpServer

import (
	"github.com/gin-gonic/gin"
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

func authJwtMiddleware(env *Environment) gin.HandlerFunc {
	return func(c *gin.Context) {
		// extract jwt token from authorization header if present
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) < 1 {
			c.Next()
			return
		}

		if user, err := checkToken(tokenStr, env.Authentication.JwtSecret()); err != nil {
			jsonErrorResponse(c, http.StatusUnauthorized, errors.New("invalid jwt token"))
			c.Abort()
		} else {
			// continue; if user is set this means a valid token is present
			c.Set("AuthUser", user)
			c.Next()
		}
	}
}

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

func isViewAuthenticated(view ViewConfig, c *gin.Context, allowAnonymous bool) bool {
	return isViewAuthenticatedByUser(view, c.GetString("AuthUser"), allowAnonymous)
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
