// Command repair-cross-unit-oversell finds Oversell settlements that served a
// Line from a Stock of another Unit (ADR-0002 forbids them) and, with -apply,
// undoes each one through the Stock ledger: the quantity goes back to the
// Stock and the Line is owed it again. It reports by default. It reads
// MONGO_HOST and MONGO_POS_DB_NAME from the environment, and creates no
// indexes. Settlements recorded with the old "ADJUST:" marker name no Stock,
// so they cannot be checked and are left alone. A Line put back into debt is
// served by the next Stock of its own Unit to arrive (ADR-0002), not from what
// that Unit already holds; Lines with no Unit are never touched.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"pos/app/data/ledger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	apply := flag.Bool("apply", false, "undo the cross-Unit settlements")
	by := flag.String("by", "repair-cross-unit-oversell", "recorded as the author of the history rows")
	flag.Parse()

	host, name := os.Getenv("MONGO_HOST"), os.Getenv("MONGO_POS_DB_NAME")
	if host == "" || name == "" {
		fmt.Fprintln(os.Stderr, "MONGO_HOST and MONGO_POS_DB_NAME are required")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(host))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	// The repair issues no document codes.
	l := ledger.New(client, client.Database(name), nil)
	draws, err := l.RepairCrossUnitOversell(ctx, *apply, *by)
	for _, d := range draws {
		fmt.Printf("line %s drew %d from stock %s of another Unit\n", d.Line.Hex(), d.Quantity, d.Stock.Hex())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	verb := "would undo"
	if *apply {
		verb = "undid"
	}
	fmt.Printf("%s %d cross-Unit settlements\n", verb, len(draws))
}
