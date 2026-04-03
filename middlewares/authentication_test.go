package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			return nil, errors.New("not found")
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
