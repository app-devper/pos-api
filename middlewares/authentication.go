package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

func RequireBranch(employeeEntity repositories.IEmployee, branchEntity repositories.IBranch) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.GetString("UserId")
		employee, err := employeeEntity.GetEmployeeByUserId(userId)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				errcode.Abort(ctx, http.StatusForbidden, errcode.AU_FORBIDDEN_001, "employee lookup failed")
				return
			}
			defaultBranch, bErr := branchEntity.GetBranchByCode("HQ")
			if bErr != nil {
				errcode.Abort(ctx, http.StatusForbidden, errcode.AU_FORBIDDEN_001, "no branch available")
				return
			}
			ctx.Set("BranchId", defaultBranch.Id.Hex())
			ctx.Set("EmployeeRole", "STAFF")
		} else {
			if employee.Status != "" && employee.Status != constant.ACTIVE {
				errcode.Abort(ctx, http.StatusForbidden, errcode.AU_FORBIDDEN_001, "employee inactive")
				return
			}
			ctx.Set("BranchId", employee.BranchId.Hex())
			ctx.Set("EmployeeRole", employee.Role)
		}
		ctx.Next()
	}
}

type AccessClaims struct {
	Role     string `json:"role"`
	System   string `json:"system"`
	ClientId string `json:"clientId"`
	jwt.RegisteredClaims
}

type authConfig struct {
	jwtKey   []byte
	clientId string
	system   string
}

func RequireAuthenticated() gin.HandlerFunc {
	config, configErr := loadAuthConfig()
	return func(ctx *gin.Context) {
		if configErr != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.SY_INTERNAL_001, configErr.Error())
			return
		}

		token := ctx.GetHeader("Authorization")
		if token == "" {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_001, "missing authorization header")
			return
		}
		jwtToken := strings.Split(token, "Bearer ")
		if len(jwtToken) < 2 {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_001, "missing authorization header")
			return
		}
		claims := &AccessClaims{}
		tkn, err := jwt.ParseWithClaims(jwtToken[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.jwtKey, nil
		})
		if err != nil {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_002, err.Error())
			return
		}
		if tkn == nil || !tkn.Valid || claims.ID == "" {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_002, "token invalid")
			return
		}
		if config.system != claims.System {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_003, "system invalid")
			return
		}
		if config.clientId != claims.ClientId {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_004, "clientId invalid")
			return
		}

		ctx.Set("SessionId", claims.ID)
		ctx.Set("Role", claims.Role)
		ctx.Set("System", claims.System)
		ctx.Set("ClientId", claims.ClientId)
		ctx.Next()
	}
}

func RequireSession(sessionEntity repositories.ISession) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId := ctx.GetString("SessionId")
		userId, err := sessionEntity.GetSessionById(sessionId)
		if err != nil {
			errcode.Abort(ctx, http.StatusUnauthorized, errcode.AU_UNAUTHORIZED_005, "session invalid")
			return
		}
		ctx.Set("UserId", userId)
		ctx.Next()
	}
}

func loadAuthConfig() (*authConfig, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("missing required env: SECRET_KEY")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		return nil, errors.New("missing required env: CLIENT_ID")
	}

	system := os.Getenv("SYSTEM")
	if system == "" {
		return nil, errors.New("missing required env: SYSTEM")
	}

	return &authConfig{
		jwtKey:   []byte(secretKey),
		clientId: clientId,
		system:   system,
	}, nil
}
