package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type employeeRepoStub struct {
	repositories.IEmployee
	getByUserIDFn func(userId string) (*entities.Employee, error)
}

func (s *employeeRepoStub) GetEmployeeByUserId(userId string) (*entities.Employee, error) {
	return s.getByUserIDFn(userId)
}

type branchRepoStub struct {
	repositories.IBranch
	getByCodeFn func(code string) (*entities.Branch, error)
}

func (s *branchRepoStub) GetBranchByCode(code string) (*entities.Branch, error) {
	return s.getByCodeFn(code)
}

func TestRequireBranchUsesEmployeeBranchAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID()
	employeeRepo := &employeeRepoStub{
		getByUserIDFn: func(userId string) (*entities.Employee, error) {
			return &entities.Employee{BranchId: branchID, Role: "ADMIN"}, nil
		},
	}
	branchRepo := &branchRepoStub{
		getByCodeFn: func(code string) (*entities.Branch, error) {
			t.Fatalf("did not expect HQ fallback lookup")
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "user-1")

	called := false
	handler := RequireBranch(employeeRepo, branchRepo)
	handler(ctx)
	if !ctx.IsAborted() {
		called = true
	}

	if !called {
		t.Fatalf("expected middleware to continue")
	}
	if got := ctx.GetString("BranchId"); got != branchID.Hex() {
		t.Fatalf("expected BranchId %s, got %s", branchID.Hex(), got)
	}
	if got := ctx.GetString("EmployeeRole"); got != "ADMIN" {
		t.Fatalf("expected EmployeeRole ADMIN, got %s", got)
	}
}

func TestRequireBranchFallsBackToHQStaffWhenEmployeeMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hqID := primitive.NewObjectID()
	employeeRepo := &employeeRepoStub{
		getByUserIDFn: func(userId string) (*entities.Employee, error) {
			return nil, mongo.ErrNoDocuments
		},
	}
	branchRepo := &branchRepoStub{
		getByCodeFn: func(code string) (*entities.Branch, error) {
			if code != "HQ" {
				t.Fatalf("expected HQ lookup, got %s", code)
			}
			return &entities.Branch{Id: hqID, Code: "HQ"}, nil
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "user-1")

	RequireBranch(employeeRepo, branchRepo)(ctx)

	if ctx.IsAborted() {
		t.Fatalf("expected middleware to continue")
	}
	if got := ctx.GetString("BranchId"); got != hqID.Hex() {
		t.Fatalf("expected BranchId %s, got %s", hqID.Hex(), got)
	}
	if got := ctx.GetString("EmployeeRole"); got != "STAFF" {
		t.Fatalf("expected EmployeeRole STAFF, got %s", got)
	}
}

func TestRequireBranchFailsWhenNoFallbackBranchExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeRepo := &employeeRepoStub{
		getByUserIDFn: func(userId string) (*entities.Employee, error) {
			return nil, errors.New("not found")
		},
	}
	branchRepo := &branchRepoStub{
		getByCodeFn: func(code string) (*entities.Branch, error) {
			return nil, errors.New("missing HQ")
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "user-1")

	RequireBranch(employeeRepo, branchRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if body := w.Body.String(); body == "" || !strings.Contains(body, errcode.AU_FORBIDDEN_001) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.AU_FORBIDDEN_001, body)
	}
}

func TestRequireBranchFailsWhenEmployeeLookupReturnsSystemError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeRepo := &employeeRepoStub{
		getByUserIDFn: func(userId string) (*entities.Employee, error) {
			return nil, errors.New("database timeout")
		},
	}
	branchRepo := &branchRepoStub{
		getByCodeFn: func(code string) (*entities.Branch, error) {
			t.Fatalf("did not expect HQ fallback lookup")
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "user-1")

	RequireBranch(employeeRepo, branchRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "employee lookup failed") {
		t.Fatalf("expected employee lookup failure response, got %s", body)
	}
}

func TestRequireBranchStillFallsBackWhenEmployeeMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hqID := primitive.NewObjectID()
	employeeRepo := &employeeRepoStub{
		getByUserIDFn: func(userId string) (*entities.Employee, error) {
			return nil, mongo.ErrNoDocuments
		},
	}
	branchRepo := &branchRepoStub{
		getByCodeFn: func(code string) (*entities.Branch, error) {
			return &entities.Branch{Id: hqID, Code: "HQ"}, nil
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "user-1")

	RequireBranch(employeeRepo, branchRepo)(ctx)

	if ctx.IsAborted() {
		t.Fatalf("expected middleware to continue")
	}
	if got := ctx.GetString("BranchId"); got != hqID.Hex() {
		t.Fatalf("expected BranchId %s, got %s", hqID.Hex(), got)
	}
}

func TestRequireAuthenticatedRejectsMissingAuthConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	restore := snapshotEnv([]string{"SECRET_KEY", "CLIENT_ID", "SYSTEM"})
	defer restore()

	t.Setenv("SECRET_KEY", "")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("SYSTEM", "pos")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.Header.Set("Authorization", "Bearer token")

	RequireAuthenticated()(ctx)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, errcode.SY_INTERNAL_001) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.SY_INTERNAL_001, body)
	}
}

func TestRequireAuthenticatedAcceptsValidTokenWhenConfigPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	restore := snapshotEnv([]string{"SECRET_KEY", "CLIENT_ID", "SYSTEM"})
	defer restore()

	t.Setenv("SECRET_KEY", "super-secret")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("SYSTEM", "pos")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims{
		Role:     "ADMIN",
		System:   "pos",
		ClientId: "client",
		RegisteredClaims: jwt.RegisteredClaims{
			ID: "session-1",
		},
	})
	signedToken, err := token.SignedString([]byte("super-secret"))
	if err != nil {
		t.Fatalf("expected signed token, got %v", err)
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+signedToken)

	RequireAuthenticated()(ctx)

	if ctx.IsAborted() {
		t.Fatalf("expected middleware to continue, got status %d body %s", w.Code, w.Body.String())
	}
	if got := ctx.GetString("SessionId"); got != "session-1" {
		t.Fatalf("expected SessionId session-1, got %s", got)
	}
}

func snapshotEnv(keys []string) func() {
	values := make(map[string]*string, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			copied := value
			values[key] = &copied
			continue
		}
		values[key] = nil
	}

	return func() {
		for _, key := range keys {
			if values[key] == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *values[key])
		}
	}
}
