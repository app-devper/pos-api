package repositories

import "testing"

func TestSessionPrefixMatchesUmApiSessionKeyFormat(t *testing.T) {
	if sessionPrefix != "session:" {
		t.Fatalf("sessionPrefix = %q, want %q", sessionPrefix, "session:")
	}
}

func TestParseSessionUserId(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"valid session document", `{"userId":"user-1","createdAt":"2026-01-01T00:00:00Z"}`, "user-1", false},
		{"document missing userId field", `{"createdAt":"2026-01-01T00:00:00Z"}`, "", false},
		{"malformed json", `not-json`, "", true},
		{"empty payload", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSessionUserId([]byte(tc.raw))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
