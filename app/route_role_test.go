package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain"
	"pos/app/domain/constant"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Every route's Policy is exercised with a USER, a MANAGER and an ADMIN
// session, driven by docs/route-policies.txt so a new route is covered as
// soon as the manifest lists it. Only the gate is under test: a handler that
// answers 400 or 404 for a dummy body or id has been let through.

const (
	roleTestSecret = "test-secret"
	roleTestSystem = "POS"
	roleTestClient = "000"
	roleTestUser   = "user-1"
	// reachedHandler is what a request that got past its Policy answers
	// when its handler then touches a repository this router leaves nil.
	reachedHandler = 599
	// featureCode is the package path every route handler lives under; a
	// panic with no frame there happened inside a Policy.
	featureCode = "pos/app/featues/"
)

// minimumRole is, per Policy, the lowest UM role it admits.
var minimumRole = map[string]sessionclient.Role{
	"staff":        sessionclient.RoleUser,
	"signed-in":    sessionclient.RoleUser,
	"branch-admin": sessionclient.RoleAdmin,
	"shop-admin":   sessionclient.RoleAdmin,
}

// liveSessions confirms every session for roleTestUser.
type liveSessions struct{}

func (liveSessions) Session(context.Context, string) (sessionclient.Session, error) {
	return sessionclient.Session{UserId: roleTestUser, System: roleTestSystem}, nil
}

// activeEmployee is the caller's employee record: active, in a branch.
type activeEmployee struct {
	repositories.IEmployee
	branch primitive.ObjectID
}

func (e activeEmployee) GetEmployeeByUserId(userId string) (*entities.Employee, error) {
	return &entities.Employee{UserId: userId, BranchId: e.branch, Status: constant.ACTIVE, Role: "STAFF"}, nil
}

// buildSignedInRouter is buildRouter with a session store and an employee
// record that let a signed request reach each route's Policy. Every other
// repository stays nil: a handler that touches one panics, and
// reachedHandlerOnPanic turns that into reachedHandler.
func buildSignedInRouter(t *testing.T) *gin.Engine {
	t.Helper()
	t.Setenv("SECRET_KEY", roleTestSecret)
	t.Setenv("CLIENT_ID", roleTestClient)
	t.Setenv("SYSTEM", roleTestSystem)
	gin.SetMode(gin.TestMode)
	auth, err := middlewares.NewAuthWithStore(liveSessions{})
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(reachedHandlerOnPanic(t))
	applyFeatureAPIs(r.Group("/api/pos/v1"), &domain.Repository{Auth: auth, Employee: activeEmployee{branch: primitive.NewObjectID()}})
	return r
}

// reachedHandlerOnPanic answers reachedHandler when a route handler panics,
// and fails the test when the panic came from inside a Policy instead.
func reachedHandlerOnPanic(t *testing.T) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if stack := string(debug.Stack()); !strings.Contains(stack, featureCode) {
				t.Errorf("%s %s: panic inside its Policy: %v\n%s", ctx.Request.Method, ctx.FullPath(), v, stack)
			}
			ctx.AbortWithStatus(reachedHandler)
		}()
		ctx.Next()
	}
}

// accessToken signs a UM access token (HS256) for role, as UM would.
func accessToken(t *testing.T, role sessionclient.Role) string {
	t.Helper()
	enc := base64.RawURLEncoding
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := json.Marshal(map[string]any{"role": role, "system": roleTestSystem, "clientId": roleTestClient,
		"jti": "session-" + string(role), "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	unsigned := enc.EncodeToString(header) + "." + enc.EncodeToString(claims)
	mac := hmac.New(sha256.New, []byte(roleTestSecret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc.EncodeToString(mac.Sum(nil))
}

type policyRoute struct{ method, path, policy string }

func manifestRoutes(t *testing.T) []policyRoute {
	t.Helper()
	body, err := os.ReadFile(policyManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var routes []policyRoute
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		f := strings.Fields(line)
		if len(f) != 3 {
			t.Fatalf("manifest line %q is not METHOD PATH POLICY", line)
		}
		routes = append(routes, policyRoute{f[0], f[1], f[2]})
	}
	return routes
}

func callAs(r *gin.Engine, route policyRoute, token string) int {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(route.method, requestPath(route.path), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(recorder, req)
	return recorder.Code
}

func TestEveryRouteAdmitsExactlyTheRolesItsPolicyNames(t *testing.T) {
	r := buildSignedInRouter(t)
	roles := []sessionclient.Role{sessionclient.RoleUser, sessionclient.RoleManager, sessionclient.RoleAdmin}
	tokens := map[sessionclient.Role]string{}
	for _, role := range roles {
		tokens[role] = accessToken(t, role)
	}
	checked := map[string]int{}
	for _, route := range manifestRoutes(t) {
		min, known := minimumRole[route.policy]
		if !known {
			t.Errorf("%s %s: policy %q has no expected minimum role", route.method, route.path, route.policy)
			continue
		}
		for _, role := range roles {
			got := callAs(r, route, tokens[role])
			if role.AtLeast(min) && (got == http.StatusUnauthorized || got == http.StatusForbidden) {
				t.Errorf("%s %s (%s): %s refused with %d, want let through", route.method, route.path, route.policy, role, got)
			}
			if !role.AtLeast(min) && got != http.StatusForbidden {
				t.Errorf("%s %s (%s): %s answered %d, want 403", route.method, route.path, route.policy, role, got)
			}
		}
		checked[route.policy]++
	}
	if len(checked) != len(minimumRole) {
		t.Fatalf("checked policies %v, want all of %v", checked, minimumRole)
	}
	t.Logf("checked routes per policy: %v", fmt.Sprint(checked))
}
