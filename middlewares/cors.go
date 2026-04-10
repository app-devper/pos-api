package middlewares

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCors return new gin handler fuc to handle CORS request.
// When allowedOrigins contains only "*", AllowCredentials is disabled because
// browsers reject credentialed requests to a wildcard origin (CORS spec).
func NewCors(allowedOrigins []string) gin.HandlerFunc {
	isWildcard := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{
			"Origin", "Host",
			"Content-Type", "Content-Length",
			"Accept-Encoding", "Accept-Language", "Accept",
			"X-CSRF-Token", "Authorization", "X-Requested-With", "X-Access-Token",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: !isWildcard,
	})
}
