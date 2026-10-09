package app

import (
	"os"
	"strings"
	"testing"

	"pos/middlewares"
)

// policyManifestPath lists who may call every route. Reviewing a change to it
// is how a route's policy changes on purpose: a write that should need an
// Admin and does not shows up here as "staff".
const policyManifestPath = "../docs/route-policies.txt"

func TestEveryRouteIsRegisteredThroughAPolicy(t *testing.T) {
	r := buildRouter(t)
	byRoute := map[string]bool{}
	for _, line := range middlewares.RoutePolicies() {
		f := strings.Fields(line)
		byRoute[f[0]+" "+f[1]] = true
	}
	for _, route := range r.Routes() {
		if !strings.HasPrefix(route.Path, "/api/pos/v1") {
			continue
		}
		if !byRoute[route.Method+" "+route.Path] {
			t.Errorf("%s %s is registered without a Policy", route.Method, route.Path)
		}
	}
}

func TestRoutePolicyManifestMatchesTheRegisteredRoutes(t *testing.T) {
	buildRouter(t)
	want := strings.Join(middlewares.RoutePolicies(), "\n") + "\n"

	if os.Getenv("UPDATE_ROUTE_MANIFEST") != "" {
		if err := os.WriteFile(policyManifestPath, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile(policyManifestPath)
	if err != nil {
		t.Fatalf("reading %s: %v\n\nRegenerate it with:\n  UPDATE_ROUTE_MANIFEST=1 go test ./app -run Manifest", policyManifestPath, err)
	}
	if string(got) != want {
		t.Errorf("%s is out of date. Regenerate it with:\n  UPDATE_ROUTE_MANIFEST=1 go test ./app -run Manifest\nand review who may now call what.\n\n%s",
			policyManifestPath, diffLines(string(got), want))
	}
}
