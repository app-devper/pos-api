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

func TestRun_BadFlagExitsTwo(t *testing.T) {
	c := &call{}
	code, _, _ := exec([]string{"-nope"}, fullEnv, fakeMigrate(c, legacy.Report{}, nil))
	if code != 2 || c.count != 0 {
		t.Fatalf("code=%d calls=%d, want 2 and 0", code, c.count)
	}
}
