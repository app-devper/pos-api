package middlewares

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"pos/app/data/repositories"
	"strings"
)

type AccessClaims struct {
	Role     string `json:"role"`
	System   string `json:"system"`
	ClientId string `json:"clientId"`
	jwt.StandardClaims
}

func RequireAuthenticated() gin.HandlerFunc {
	jwtKey := []byte(os.Getenv("SECRET_KEY"))
	clientId := os.Getenv("CLIENT_ID")
	system := os.Getenv("SYSTEM")
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		jwtToken := strings.Split(token, "Bearer ")
		if len(jwtToken) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		claims := &AccessClaims{}
		tkn, err := jwt.ParseWithClaims(jwtToken[1], claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if tkn == nil || !tkn.Valid || claims.Id == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
			return
		}
		if system != claims.System {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "system invalid"})
			return
		}
		if clientId != claims.ClientId {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "clientId invalid"})
			return
		}

		ctx.Set("SessionId", claims.Id)
		ctx.Set("Role", claims.Role)
		ctx.Set("System", claims.System)
		ctx.Set("ClientId", claims.ClientId)

		logrus.Info("SessionId: " + claims.Id)
		logrus.Info("Role: " + claims.Role)
		logrus.Info("System: " + claims.System)
		logrus.Info("ClientId: " + claims.ClientId)
		return
	}
}

func RequireSession(sessionEntity repositories.ISession) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId := ctx.GetString("SessionId")
		userId, err := sessionEntity.GetSessionById(sessionId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session invalid"})
			return
		}
		ctx.Set("UserId", userId)
		logrus.Info("UserId: " + userId)
		return
	}
}
