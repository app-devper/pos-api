package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/entities"
	"pos/app/domain/constant"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func runActive(t *testing.T, employee *entities.Employee, err error) (int, bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("UserId", "u1")
	reached := false
	repo := &employeeRepoStub{getByUserIDFn: func(string) (*entities.Employee, error) { return employee, err }}
	h := RequireActiveEmployee(repo)
	h(ctx)
	if !ctx.IsAborted() {
		reached = true
	}
	return w.Code, reached
}

func TestRequireActiveEmployee(t *testing.T) {
	cases := []struct {
		name     string
		employee *entities.Employee
		err      error
		passes   bool
	}{
		{"an active employee passes", &entities.Employee{Status: constant.ACTIVE}, nil, true},
		{"a record with no status passes, as RequireBranch allows", &entities.Employee{}, nil, true},
		{"no employee record passes: shop-wide routes need none", nil, mongo.ErrNoDocuments, true},
		{"an inactive employee is refused", &entities.Employee{Status: "INACTIVE"}, nil, false},
		{"a failed lookup is refused", nil, errors.New("mongo down"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, passed := runActive(t, c.employee, c.err)
			if passed != c.passes {
				t.Fatalf("passed=%v (status %d), want %v", passed, code, c.passes)
			}
			if !c.passes && code != http.StatusForbidden {
				t.Fatalf("status %d, want 403", code)
			}
		})
	}
}
