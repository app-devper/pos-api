// Command repair-order-totals recomputes the money of Orders whose total was
// overwritten wrongly when one of their Lines was cancelled. It reports by
// default and writes only with -apply. It reads MONGO_HOST and
// MONGO_POS_DB_NAME from the environment, and creates no indexes.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"pos/app/data/repositories"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	apply := flag.Bool("apply", false, "write the recomputed totals")
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

	repairs, err := repositories.RepairCancelledLineTotals(ctx, client.Database(name), *apply)
	for _, r := range repairs {
		fmt.Printf("%s %s total %.2f -> %.2f, totalCost %.2f -> %.2f, discount %.2f -> %.2f\n",
			r.OrderId.Hex(), r.Code, r.Total, r.NewTotal, r.TotalCost, r.NewTotalCost, r.Discount, r.NewDiscount)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	verb := "would change"
	if *apply {
		verb = "changed"
	}
	fmt.Printf("%s %d orders\n", verb, len(repairs))
}
