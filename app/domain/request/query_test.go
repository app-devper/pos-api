package request

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBoolQueryReadsAnOptInFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name      string
		url       string
		wantValue bool
		wantOk    bool
	}{
		{"absent", "/x", false, true},
		{"true", "/x?flag=true", true, true},
		{"TRUE", "/x?flag=TRUE", true, true},
		{"false", "/x?flag=false", false, true},
		{"False", "/x?flag=False", false, true},
		{"empty", "/x?flag=", false, false},
		{"maybe", "/x?flag=maybe", false, false},
		{"one", "/x?flag=1", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", tc.url, nil)

			value, ok := BoolQuery(ctx, "flag")

			if value != tc.wantValue || ok != tc.wantOk {
				t.Fatalf("BoolQuery(%q) = %v, %v; want %v, %v", tc.url, value, ok, tc.wantValue, tc.wantOk)
			}
		})
	}
}
