package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"pos/app/domain"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

// buildRouter registers every feature the same way StartGin does. The handlers
// never run here: RequireSession rejects a request with no Authorization
// header before anything reaches a repository or the session store, so the
// nil fields are never dereferenced.
func buildRouter(t *testing.T) *gin.Engine {
	t.Helper()
	t.Setenv("SECRET_KEY", "test-secret")
	t.Setenv("CLIENT_ID", "000")
	t.Setenv("SYSTEM", "POS")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", healthCheck())
	r.GET("/api/pos/health", healthCheck())
	auth, err := middlewares.NewAuthWithStore(unusedStore{})
	if err != nil {
		t.Fatal(err)
	}
	applyFeatureAPIs(r.Group("/api/pos/v1"), &domain.Repository{Auth: auth})
	return r
}

// unusedStore stands in for UM's session store; anonymous requests never
// reach it.
type unusedStore struct{}

func (unusedStore) Session(context.Context, string) (sessionclient.Session, error) {
	return sessionclient.Session{}, sessionclient.ErrUnavailable
}

var pathParam = regexp.MustCompile(`:[a-zA-Z]+`)

// requestPath turns a registered pattern such as "/products/:id" into
// something routable.
func requestPath(pattern string) string {
	return pathParam.ReplaceAllString(pattern, "sample")
}

func TestEveryBusinessRouteRejectsAnAnonymousRequest(t *testing.T) {
	r := buildRouter(t)

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatal("no routes registered")
	}

	checked := 0
	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/api/pos/v1") {
			continue
		}
		checked++

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(route.Method, requestPath(route.Path), nil)
		r.ServeHTTP(recorder, req)

		// A route that answers anything else has lost its auth middleware, or
		// is failing for a reason a caller should not be able to trigger
		// without credentials.
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401 without a token, got %d",
				route.Method, route.Path, recorder.Code)
		}
	}

	if checked == 0 {
		t.Fatal("no /api/pos/v1 routes were checked")
	}
	t.Logf("checked %d business routes", checked)
}

func TestHealthRoutesStayPublic(t *testing.T) {
	r := buildRouter(t)

	for _, path := range []string{"/health", "/api/pos/health"} {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusOK {
			t.Errorf("%s: expected 200 without a token, got %d", path, recorder.Code)
		}
	}
}

func TestABearerTokenThatIsNotOursIsRejected(t *testing.T) {
	r := buildRouter(t)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pos/v1/products", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for a malformed token, got %d", recorder.Code)
	}
}
