package middlewares

import (
	"errors"
	"net/http"
	"os"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/constant"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/app-devper/um-api/sessionclient/ginauth"
	"github.com/gin-gonic/gin"
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

// NewAuth verifies UM access tokens for pos: SYSTEM and CLIENT_ID pin the
// token, and the session is confirmed in UM's Redis at redisHost
// (um-api ADR-0005). It fails when any of them is missing.
func NewAuth(redisHost string) (*ginauth.Auth, error) {
	store, err := sessionclient.RedisStoreFor(redisHost)
	if err != nil {
		return nil, err
	}
	return NewAuthWithStore(store)
}

// NewAuthWithStore is NewAuth with UM's session store supplied, for tests.
func NewAuthWithStore(store sessionclient.Store) (*ginauth.Auth, error) {
	verifier, err := sessionclient.NewVerifier(sessionclient.Config{
		SecretKey: os.Getenv("SECRET_KEY"),
		System:    os.Getenv("SYSTEM"),
		ClientID:  os.Getenv("CLIENT_ID"),
		Store:     store,
	})
	if err != nil {
		return nil, err
	}
	return ginauth.New(verifier, func(ctx *gin.Context, e *sessionclient.Error) {
		errcode.Abort(ctx, e.Status, e.Code, e.Message)
	}), nil
}

// RequireSession admits a caller with a live UM session. Every pos route uses
// the default outage policy: while UM is unreachable a read may continue with
// the session last confirmed for its token, and writes wait.
func RequireSession(auth *ginauth.Auth) gin.HandlerFunc {
	return auth.Require(sessionclient.ReadOnlyWithLastGood)
}
