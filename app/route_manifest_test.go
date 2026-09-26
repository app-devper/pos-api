package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// manifestPath holds the routes this service exposes, in a form the Flutter
// client's contract test can read. It is committed on purpose: the client lives
// in another repository and its CI only ever checks out one repo, so a reviewed
// file is the only place the two sides can meet.
const manifestPath = "../docs/routes.txt"

// routeManifest renders every registered route as "METHOD PATH", sorted, with
// parameter names normalised away. The client cannot know that this service
// calls an id ":id" here and ":productId" there, and it does not need to — only
// the shape of the path matters to a caller.
func routeManifest(t *testing.T) string {
	t.Helper()

	r := buildRouter(t)
	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s", route.Method,
			pathParam.ReplaceAllString(route.Path, ":param")))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func TestRouteManifestMatchesTheRegisteredRoutes(t *testing.T) {
	want := routeManifest(t)

	if os.Getenv("UPDATE_ROUTE_MANIFEST") != "" {
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
			t.Fatalf("creating the manifest directory: %v", err)
		}
		if err := os.WriteFile(manifestPath, []byte(want), 0o644); err != nil {
			t.Fatalf("writing the manifest: %v", err)
		}
		t.Logf("wrote %s", manifestPath)
		return
	}

	got, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("reading %s: %v\n\nRegenerate it with:\n  UPDATE_ROUTE_MANIFEST=1 go test ./app -run TestRouteManifest",
			manifestPath, err)
	}

	if string(got) != want {
		t.Errorf("%s is out of date.\n\nRegenerate it with:\n  UPDATE_ROUTE_MANIFEST=1 go test ./app -run TestRouteManifest\n\nThen review the diff: a route that disappeared or changed shape breaks the Flutter client, which has no other way to find out.\n\n%s",
			manifestPath, diffLines(string(got), want))
	}
}

// diffLines reports which manifest lines were added and which went away.
func diffLines(got, want string) string {
	inGot := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(got), "\n") {
		inGot[line] = true
	}
	inWant := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(want), "\n") {
		inWant[line] = true
	}

	var b strings.Builder
	var added, removed []string
	for line := range inWant {
		if !inGot[line] {
			added = append(added, line)
		}
	}
	for line := range inGot {
		if !inWant[line] {
			removed = append(removed, line)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	for _, line := range removed {
		fmt.Fprintf(&b, "  gone from the service: %s\n", line)
	}
	for _, line := range added {
		fmt.Fprintf(&b, "  new in the service:    %s\n", line)
	}
	return b.String()
}
