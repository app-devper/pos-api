// Command migrate-legacy-db gives a restored COPY of the 2024 database the
// Branch the current API needs, and reports what a person must decide first.
// It reports by default and writes only with -apply, which must be repeated
// with -confirm-db <the database name>. It reads MONGO_HOST and
// MONGO_POS_DB_NAME from the environment, and creates no indexes. The first
// line names the database, the server's hosts (never the user, password or
// options of MONGO_HOST) and the mode.
//
// Exit codes: 0 clean, 1 error, 2 usage, 3 findings present (in apply mode the
// backfill has still happened).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"pos/app/data/legacy"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	exitClean    = 0
	exitError    = 1
	exitUsage    = 2
	exitFindings = 3

	timeout = 10 * time.Minute
)

// migrateFunc runs the migration on database name of the server at host.
type migrateFunc func(ctx context.Context, host, name string, apply bool) (legacy.Report, error)

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, connectAndMigrate))
}

func connectAndMigrate(ctx context.Context, host, name string, apply bool) (legacy.Report, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(host))
	if err != nil {
		return legacy.Report{}, fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()
	return legacy.Migrate(ctx, client.Database(name), apply)
}

// config is what one run was asked to do.
type config struct {
	uri, hosts, name string
	apply            bool
}

func run(args []string, env func(string) string, out, errOut io.Writer, migrate migrateFunc) int {
	cfg, ok := parseConfig(args, env, errOut)
	if !ok {
		return exitUsage
	}
	mode := "REPORT"
	if cfg.apply {
		mode = "APPLY"
	}
	fmt.Fprintf(out, "database %s on %s, mode %s\n", cfg.name, cfg.hosts, mode)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	report, err := migrate(ctx, cfg.uri, cfg.name, cfg.apply)
	printReport(out, report, cfg.apply)
	if err != nil {
		fmt.Fprintln(out, "\nthe run did not finish: the lines above are what was done before the error (the step that failed may have written part of its documents); running it again is safe")
		fmt.Fprintln(errOut, err)
		return exitError
	}
	if report.HasFindings() {
		if cfg.apply {
			fmt.Fprintln(out, "findings present: the backfill above HAS been written; the findings need a decision")
		}
		return exitFindings
	}
	return exitClean
}

// parseConfig reads the flags and the environment; false means a usage error,
// already explained on errOut.
func parseConfig(args []string, env func(string) string, errOut io.Writer) (config, bool) {
	flags := flag.NewFlagSet("migrate-legacy-db", flag.ContinueOnError)
	flags.SetOutput(errOut)
	apply := flags.Bool("apply", false, "write the backfill (needs -confirm-db)")
	confirm := flags.String("confirm-db", "", "with -apply: the database name, repeated")
	if err := flags.Parse(args); err != nil {
		return config{}, false
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(errOut, "unexpected argument %q (flags start with a dash, as in -apply)\n", flags.Arg(0))
		return config{}, false
	}
	uri, name := env("MONGO_HOST"), env("MONGO_POS_DB_NAME")
	if uri == "" || name == "" {
		fmt.Fprintln(errOut, "MONGO_HOST and MONGO_POS_DB_NAME are required")
		return config{}, false
	}
	hosts, ok := hostsOf(uri)
	if !ok {
		fmt.Fprintln(errOut, "MONGO_HOST is not a mongodb:// or mongodb+srv:// URI with a host (not shown, it may hold a password)")
		return config{}, false
	}
	if *apply && *confirm != name {
		fmt.Fprintf(errOut, "-apply needs -confirm-db %s (the value of MONGO_POS_DB_NAME)\n", name)
		return config{}, false
	}
	return config{uri: uri, hosts: hosts, name: name, apply: *apply}, true
}

// hostsOf returns the host list of a MongoDB URI, without user, password,
// database or options; false when uri does not parse. The parse error is
// dropped: it quotes the URI.
func hostsOf(uri string) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil || (u.Scheme != "mongodb" && u.Scheme != "mongodb+srv") || u.Host == "" {
		return "", false
	}
	return u.Host, true
}

func printReport(out io.Writer, report legacy.Report, apply bool) {
	verb := "would change"
	if apply {
		verb = "changed"
	}
	fmt.Fprintf(out, "HQ Branch %s (created: %t)\n", report.HQ.Hex(), report.HQCreated)
	for _, c := range report.Changes {
		fmt.Fprintf(out, "%s %s: %s %d\n", c.Collection, c.Field, verb, c.Count)
	}
	for _, f := range report.Findings {
		fmt.Fprintf(out, "\nFINDING %s (%d): %s\n", f.Kind, f.Total, f.Detail)
		for _, id := range f.Ids {
			fmt.Fprintf(out, "  %s\n", id)
		}
		if f.Total > int64(len(f.Ids)) {
			fmt.Fprintf(out, "  ... and %d more\n", f.Total-int64(len(f.Ids)))
		}
	}
}
