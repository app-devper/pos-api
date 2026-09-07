package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckRespondsOkWithoutAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", healthCheck())
	r.GET("/api/pos/health", healthCheck())

	for _, path := range []string{"/health", "/api/pos/health"} {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
			t.Fatalf("%s: expected ok status, got %s", path, recorder.Body.String())
		}
	}
}

func TestHealthCheckReportsCloudRunServiceName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("K_SERVICE", "pos-dev-api")

	r := gin.New()
	r.GET("/health", healthCheck())

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if !strings.Contains(recorder.Body.String(), `"service":"pos-dev-api"`) {
		t.Fatalf("expected service name from K_SERVICE, got %s", recorder.Body.String())
	}
}
