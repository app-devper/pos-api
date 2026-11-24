package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"pos/app/data/legacy"
)

type call struct {
	count int
	host  string
	name  string
	apply bool
}

func fakeMigrate(c *call, report legacy.Report, err error) migrateFunc {
	return func(_ context.Context, host, name string, apply bool) (legacy.Report, error) {
		c.count++
		c.host, c.name, c.apply = host, name, apply
		return report, err
	}
}

func envOf(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

var fullEnv = map[string]string{"MONGO_HOST": "mongodb://example", "MONGO_POS_DB_NAME": "copy_db"}

func exec(args []string, env map[string]string, m migrateFunc) (int, string, string) {
	var out, errOut bytes.Buffer
	code := run(args, envOf(env), &out, &errOut, m)
	return code, out.String(), errOut.String()
}

func TestRun_NeedsHostAndName(t *testing.T) {
	// Arrange
	c := &call{}
	// Act
	code, _, errOut := exec(nil, map[string]string{}, fakeMigrate(c, legacy.Report{}, nil))
	// Assert
	if code != 2 || c.count != 0 {
		t.Fatalf("code=%d calls=%d, want 2 and 0", code, c.count)
	}
	for _, want := range []string{"MONGO_HOST", "MONGO_POS_DB_NAME"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %s", errOut, want)
		}
	}
}

func TestRun_ApplyNeedsConfirmDb(t *testing.T) {
	for _, args := range [][]string{{"-apply"}, {"-apply", "-confirm-db", "other_db"}} {
		c := &call{}
		code, _, errOut := exec(args, fullEnv, fakeMigrate(c, legacy.Report{}, nil))
		if code != 2 || c.count != 0 {
			t.Errorf("args %v: code=%d calls=%d, want 2 and 0", args, code, c.count)
		}
		if !strings.Contains(errOut, "-confirm-db") {
			t.Errorf("args %v: stderr %q does not mention -confirm-db", args, errOut)
		}
	}
}

func TestRun_ReportModeCallsMigrateWithoutApply(t *testing.T) {
	c := &call{}
	report := legacy.Report{Changes: []legacy.Change{{Collection: "orders", Field: "branchId", Count: 7}}}
	code, out, _ := exec(nil, fullEnv, fakeMigrate(c, report, nil))
	if code != 0 || c.count != 1 || c.apply {
		t.Fatalf("code=%d calls=%d apply=%v, want 0, 1, false", code, c.count, c.apply)
	}
	if c.host != "mongodb://example" || c.name != "copy_db" {
		t.Errorf("migrate got host=%q name=%q", c.host, c.name)
	}
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[0], "copy_db") || !strings.Contains(lines[0], "REPORT") {
		t.Errorf("first line %q lacks database name and mode", lines[0])
	}
	if !strings.Contains(out, "orders branchId: would change 7") {
		t.Errorf("output lacks the change line:\n%s", out)
	}
}

func TestRun_ApplyWithMatchingConfirmWrites(t *testing.T) {
	c := &call{}
	report := legacy.Report{Changes: []legacy.Change{{Collection: "orders", Field: "branchId", Count: 7}}}
	code, out, _ := exec([]string{"-apply", "-confirm-db", "copy_db"}, fullEnv, fakeMigrate(c, report, nil))
	if code != 0 || c.count != 1 || !c.apply {
		t.Fatalf("code=%d calls=%d apply=%v, want 0, 1, true", code, c.count, c.apply)
	}
	if !strings.Contains(strings.Split(out, "\n")[0], "APPLY") || !strings.Contains(out, "orders branchId: changed 7") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestRun_FindingsExitThree(t *testing.T) {
	report := legacy.Report{Findings: []legacy.Finding{
		{Kind: legacy.KindNegativeStock, Detail: "Stock below zero", Ids: []string{"a1", "b2"}, Total: 45},
	}}
	for _, args := range [][]string{nil, {"-apply", "-confirm-db", "copy_db"}} {
		c := &call{}
		code, out, _ := exec(args, fullEnv, fakeMigrate(c, report, nil))
		if code != 3 {
			t.Fatalf("args %v: code=%d, want 3", args, code)
		}
		for _, want := range []string{legacy.KindNegativeStock, "45", "a1", "b2"} {
			if !strings.Contains(out, want) {
				t.Errorf("args %v: output lacks %q:\n%s", args, want, out)
			}
		}
		if len(args) > 0 && !strings.Contains(out, "backfill") {
			t.Errorf("apply output must say the backfill happened:\n%s", out)
		}
	}
}

func TestRun_MigrateErrorExitsOne(t *testing.T) {
	c := &call{}
	code, _, errOut := exec(nil, fullEnv, fakeMigrate(c, legacy.Report{}, errors.New("boom")))
	if code != 1 || !strings.Contains(errOut, "boom") {
		t.Fatalf("code=%d stderr=%q, want 1 and boom", code, errOut)
	}
}

func TestRun_FailedApplyPrintsWhatWasWrittenBeforeTheError(t *testing.T) {
	// Arrange
	c := &call{}
	partial := legacy.Report{
		Reached:  legacy.StageWriting,
		Changes:  []legacy.Change{{Collection: "product_stocks", Field: "branchId", Count: 4}},
		Findings: []legacy.Finding{{Kind: legacy.KindNegativeStock, Detail: "Stock below zero", Ids: []string{"a1"}, Total: 1}},
	}
	// Act
	code, out, errOut := exec([]string{"-apply", "-confirm-db", "copy_db"}, fullEnv, fakeMigrate(c, partial, errors.New("set branchId on orders: boom")))
	// Assert
	if code != 1 || !strings.Contains(errOut, "boom") {
		t.Fatalf("code=%d stderr=%q, want 1 and the error", code, errOut)
	}
	for _, want := range []string{"product_stocks branchId: changed 4", legacy.KindNegativeStock, "did not finish", "may have written part", "again is safe"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRun_FirstLineShowsTheHostsWithoutCredentials(t *testing.T) {
	// Arrange
	env := map[string]string{
		"MONGO_HOST":        "mongodb://alice:s3cretpw@db1.example:27017,db2.example:27018/admin?replicaSet=rs0&authSource=admin",
		"MONGO_POS_DB_NAME": "copy_db",
	}
	for _, migrateErr := range []error{nil, errors.New("boom")} {
		c := &call{}
		// Act
		_, out, errOut := exec([]string{"-apply", "-confirm-db", "copy_db"}, env, fakeMigrate(c, legacy.Report{}, migrateErr))
		// Assert
		first := strings.Split(out, "\n")[0]
		for _, want := range []string{"db1.example:27017,db2.example:27018", "copy_db", "APPLY"} {
			if !strings.Contains(first, want) {
				t.Errorf("err %v: first line %q lacks %q", migrateErr, first, want)
			}
		}
		for _, secret := range []string{"alice", "s3cretpw", "replicaSet", "authSource"} {
			if strings.Contains(out+errOut, secret) {
				t.Errorf("err %v: output shows %q:\n%s%s", migrateErr, secret, out, errOut)
			}
		}
	}
}

func TestRun_UnparsableHostIsUsageErrorThatDoesNotEchoIt(t *testing.T) {
	for _, uri := range []string{"mongodb://alice:s3cretpw@db1:notaport/", "alice:s3cretpw@db1", "mongodb://alice:s3cretpw@/x",
		"mongodb://alice:s3cretpw@h1:27017,h2:x/db", "mongodb+srv://alice:s3cretpw@h1,h2/db"} {
		// Arrange
		c := &call{}
		env := map[string]string{"MONGO_HOST": uri, "MONGO_POS_DB_NAME": "copy_db"}
		// Act
		code, out, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))
		// Assert
		if code != 2 || c.count != 0 {
			t.Errorf("%s: code=%d calls=%d, want 2 and 0", uri, code, c.count)
		}
		if !strings.Contains(errOut, "MONGO_HOST") || strings.Contains(out+errOut, "s3cretpw") || strings.Contains(out+errOut, "alice") {
			t.Errorf("%s: output must name MONGO_HOST without echoing it:\n%s%s", uri, out, errOut)
		}
	}
}

// secrets are the user and password spellings the host cases below use.
var secrets = []string{"user", "12345", "12%40345", "12@345", "12%3A345", "12:345"}

func TestRun_HostLineNeverShowsUserinfo(t *testing.T) {
	cases := []struct {
		uri, hosts string // hosts "" = usage error
	}{
		{"mongodb://user:12345?x@host/db", "host"},
		{"mongodb://user:12345#x@host/db", "host"},
		{"mongodb://user:12345/6@host/db", ""},
		{"mongodb://user?x@host", "host"},
		{"mongodb://user:12%40345@host/db", "host"},
		{"mongodb://user:12%3A345@host/db", "host"},
		{"mongodb+srv://user:12345@cluster0.example.net/?retryWrites=true", "cluster0.example.net"},
	}
	for _, tc := range cases {
		for _, migrateErr := range []error{nil, errors.New("boom")} {
			// Arrange
			c := &call{}
			env := map[string]string{"MONGO_HOST": tc.uri, "MONGO_POS_DB_NAME": "copy_db"}
			// Act
			code, out, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, migrateErr))
			// Assert
			for _, secret := range secrets {
				if strings.Contains(out+errOut, secret) {
					t.Errorf("%s err %v: output shows %q:\n%s%s", tc.uri, migrateErr, secret, out, errOut)
				}
			}
			if tc.hosts == "" {
				if code != 2 || c.count != 0 {
					t.Errorf("%s: code=%d calls=%d, want 2 and 0", tc.uri, code, c.count)
				}
				continue
			}
			if first := strings.Split(out, "\n")[0]; !strings.Contains(first, " on "+tc.hosts+", ") {
				t.Errorf("%s: first line %q, want hosts %q", tc.uri, first, tc.hosts)
			}
		}
	}
}

func TestRun_EveryHostOfAValidListIsAccepted(t *testing.T) {
	cases := []struct{ uri, hosts string }{
		{"mongodb://u:p@h1:27017,h2/db", "h1:27017,h2"},
		{"mongodb://%2Ftmp%2Fmongodb-27017.sock/db", "/tmp/mongodb-27017.sock"},
		{"mongodb://[::1]:27017,[::2]:1/", "[::1]:27017,[::2]:1"},
	}
	for _, tc := range cases {
		// Arrange
		c := &call{}
		env := map[string]string{"MONGO_HOST": tc.uri, "MONGO_POS_DB_NAME": "copy_db"}
		// Act
		code, out, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))
		// Assert
		if code != 0 || c.count != 1 || !strings.Contains(strings.Split(out, "\n")[0], " on "+tc.hosts+", ") {
			t.Errorf("%s: code=%d calls=%d, want 0, 1 and hosts %q:\n%s%s", tc.uri, code, c.count, tc.hosts, out, errOut)
		}
	}
}

func TestRun_AFailedRunSaysWhetherAnythingWasWritten(t *testing.T) {
	applyArgs := []string{"-apply", "-confirm-db", "copy_db"}
	cases := []struct {
		name          string
		args          []string
		reached       legacy.Stage
		want, notWant []string
	}{
		{"connect failed", nil, legacy.StageStarted, []string{"nothing was written"}, []string{"HQ Branch", "may have written"}},
		{"report mode after HQ", nil, legacy.StageHQ, []string{"HQ Branch", "nothing was written"}, []string{"may have written"}},
		{"apply before any write", applyArgs, legacy.StageFindings, []string{"nothing was written", legacy.KindNegativeStock}, []string{"HQ Branch", "may have written"}},
		{"apply after HQ, before writing", applyArgs, legacy.StageHQ, []string{"HQ Branch", "nothing was written"}, []string{"may have written"}},
		{"apply while writing", applyArgs, legacy.StageWriting, []string{"HQ Branch", "may have written part"}, []string{"nothing was written"}},
	}
	for _, tc := range cases {
		// Arrange
		c := &call{}
		report := legacy.Report{Reached: tc.reached}
		if tc.reached >= legacy.StageFindings {
			report.Findings = []legacy.Finding{{Kind: legacy.KindNegativeStock, Detail: "d", Ids: []string{"a1"}, Total: 1}}
		}
		// Act
		code, out, _ := exec(tc.args, fullEnv, fakeMigrate(c, report, errors.New("boom")))
		// Assert
		if code != 1 {
			t.Errorf("%s: code=%d, want 1", tc.name, code)
		}
		for _, want := range append([]string{"did not finish", "again is safe"}, tc.want...) {
			if !strings.Contains(out, want) {
				t.Errorf("%s: output lacks %q:\n%s", tc.name, want, out)
			}
		}
		for _, notWant := range tc.notWant {
			if strings.Contains(out, notWant) {
				t.Errorf("%s: output has %q:\n%s", tc.name, notWant, out)
			}
		}
	}
}

func TestRun_BadFlagExitsTwo(t *testing.T) {
	c := &call{}
	code, _, _ := exec([]string{"-nope"}, fullEnv, fakeMigrate(c, legacy.Report{}, nil))
	if code != 2 || c.count != 0 {
		t.Fatalf("code=%d calls=%d, want 2 and 0", code, c.count)
	}
}

func TestRun_ExtraArgumentIsUsageError(t *testing.T) {
	for _, args := range [][]string{{"apply"}, {"-apply", "-confirm-db", "copy_db", "extra"}} {
		c := &call{}
		code, _, errOut := exec(args, fullEnv, fakeMigrate(c, legacy.Report{}, nil))
		if code != 2 || c.count != 0 {
			t.Errorf("args %v: code=%d calls=%d, want 2 and 0", args, code, c.count)
		}
		if !strings.Contains(errOut, "unexpected argument") {
			t.Errorf("args %v: stderr %q does not explain", args, errOut)
		}
	}
}

func TestRun_OneVariableMissingExitsTwo(t *testing.T) {
	for _, missing := range []string{"MONGO_HOST", "MONGO_POS_DB_NAME"} {
		env := map[string]string{"MONGO_HOST": "mongodb://example", "MONGO_POS_DB_NAME": "copy_db"}
		delete(env, missing)
		c := &call{}
		code, _, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))
		if code != 2 || c.count != 0 || !strings.Contains(errOut, missing) {
			t.Errorf("missing %s: code=%d calls=%d stderr=%q", missing, code, c.count, errOut)
		}
	}
}

func TestRun_SRVOnlyOptionsAreAccepted(t *testing.T) {
	for _, uri := range []string{
		"mongodb+srv://user:12345@cluster0.example.net/?srvMaxHosts=2&retryWrites=true",
		"mongodb+srv://user:12345@cluster0.example.net/db?retryWrites=true&srvServiceName=mongo",
		"mongodb+srv://user:12345@cluster0.example.net/?SRVMAXHOSTS=2",
	} {
		// Arrange
		c := &call{}
		env := map[string]string{"MONGO_HOST": uri, "MONGO_POS_DB_NAME": "copy_db"}
		// Act
		code, out, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))
		// Assert
		if code != 0 || c.count != 1 || c.host != uri {
			t.Errorf("%s: code=%d calls=%d host passed on=%v, want 0, 1 and the URI unchanged\n%s", uri, code, c.count, c.host == uri, errOut)
		}
		if first := strings.Split(out, "\n")[0]; !strings.Contains(first, " on cluster0.example.net, ") {
			t.Errorf("%s: first line %q, want the SRV name", uri, first)
		}
		for _, secret := range secrets {
			if strings.Contains(out+errOut, secret) {
				t.Errorf("%s: output shows %q", uri, secret)
			}
		}
	}
}

func TestRun_SRVOptionOnAPlainURIIsStillRefused(t *testing.T) {
	c := &call{}
	env := map[string]string{"MONGO_HOST": "mongodb://user:12345@host/?srvMaxHosts=2", "MONGO_POS_DB_NAME": "copy_db"}

	code, _, _ := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))

	if code != 2 || c.count != 0 {
		t.Errorf("code=%d calls=%d, want 2 and 0", code, c.count)
	}
}

func TestRun_RefusedURIMessageSaysWhatToCheck(t *testing.T) {
	c := &call{}
	env := map[string]string{"MONGO_HOST": "mongodb://user:12345/6@host/db", "MONGO_POS_DB_NAME": "copy_db"}

	code, out, errOut := exec(nil, env, fakeMigrate(c, legacy.Report{}, nil))

	if code != 2 || c.count != 0 {
		t.Errorf("code=%d calls=%d, want 2 and 0", code, c.count)
	}
	for _, want := range []string{"MONGO_HOST", "driver", "percent-encoded"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("message %q lacks %q", errOut, want)
		}
	}
	if strings.Contains(errOut, "is not a mongodb://") {
		t.Errorf("message still claims the scheme is wrong: %q", errOut)
	}
	for _, secret := range secrets {
		if strings.Contains(out+errOut, secret) {
			t.Errorf("output shows %q", secret)
		}
	}
}
