package repositories

import (
	"reflect"
	"testing"
	"time"

	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var notDeleted = bson.M{"$exists": false}

// skipsDeletedProducts reports whether a pipeline drops rows whose Product has
// a deletedDate: either the products lookup filters it, or a $match after the
// lookup does.
func skipsDeletedProducts(pipeline []bson.M) bool {
	for _, stage := range pipeline {
		if match, ok := stage["$match"].(bson.M); ok && reflect.DeepEqual(match["product.deletedDate"], notDeleted) {
			return true
		}
		lookup, ok := stage["$lookup"].(bson.M)
		if !ok || lookup["from"] != "products" {
			continue
		}
		inner, _ := lookup["pipeline"].(bson.A)
		for _, s := range inner {
			if m, ok := s.(bson.M)["$match"].(bson.M); ok && reflect.DeepEqual(m["deletedDate"], notDeleted) {
				return true
			}
		}
	}
	return false
}

func TestEveryStockReportPipelineSkipsDeletedProducts(t *testing.T) {
	branch := primitive.NewObjectID().Hex()
	stockReport, err1 := buildStockReportPipeline(branch)
	lowStock, err2 := buildLowStockProductsPipeline(10, branch)
	expiring, err3 := buildExpiringProductStocksPipeline(request.GetProductLotsExpireRange{}, branch)
	deadStock, err4 := buildDeadStockProductsPipeline(time.Now(), branch)
	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			t.Fatal(err)
		}
	}
	for name, pipeline := range map[string][]bson.M{"stock report": stockReport, "low stock": lowStock, "expiring": expiring, "dead stock": deadStock} {
		if !skipsDeletedProducts(pipeline) {
			t.Errorf("%s pipeline lists deleted products", name)
		}
	}
}
