// Command migrate-legacy-db gives a restored COPY of the 2024 database the
// Branch the current API needs, and reports what a person must decide first.
// It reports by default and writes only with -apply, which must be repeated
// with -confirm-db <the database name>. It reads MONGO_HOST and
// MONGO_POS_DB_NAME from the environment, and creates no indexes.
//
// Exit codes: 0 clean, 1 error, 2 usage, 3 findings present (in apply mode the
// backfill has still happened).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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

func run(args []string, env func(string) string, out, errOut io.Writer, migrate migrateFunc) int {
	flags := flag.NewFlagSet("migrate-legacy-db", flag.ContinueOnError)
	flags.SetOutput(errOut)
	apply := flags.Bool("apply", false, "write the backfill (needs -confirm-db)")
	confirm := flags.String("confirm-db", "", "with -apply: the database name, repeated")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(errOut, "unexpected argument %q (flags start with a dash, as in -apply)\n", flags.Arg(0))
		return exitUsage
	}
	host, name := env("MONGO_HOST"), env("MONGO_POS_DB_NAME")
	if host == "" || name == "" {
		fmt.Fprintln(errOut, "MONGO_HOST and MONGO_POS_DB_NAME are required")
		return exitUsage
	}
	if *apply && *confirm != name {
		fmt.Fprintf(errOut, "-apply needs -confirm-db %s (the value of MONGO_POS_DB_NAME)\n", name)
		return exitUsage
	}

	mode := "REPORT"
	if *apply {
		mode = "APPLY"
	}
	fmt.Fprintf(out, "database %s, mode %s\n", name, mode)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	report, err := migrate(ctx, host, name, *apply)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return exitError
	}
	printReport(out, report, *apply)
	if report.HasFindings() {
		if *apply {
			fmt.Fprintln(out, "findings present: the backfill above HAS been written; the findings need a decision")
		}
		return exitFindings
	}
	return exitClean
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
