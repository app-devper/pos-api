package repositories

import (
	"testing"
	"time"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildABCAnalysisPipelineFiltersActiveOrdersByBranch(t *testing.T) {
	branchID := primitive.NewObjectID()
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	pipeline, err := buildABCAnalysisPipeline(startDate, branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	matchStage, ok := pipeline[0]["$match"].(bson.M)
	if !ok {
		t.Fatalf("expected $match stage, got %+v", pipeline[0])
	}
	if matchStage["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), matchStage["branchId"])
	}

	lookupStage, ok := pipeline[1]["$lookup"].(bson.M)
	if !ok {
		t.Fatalf("expected $lookup stage, got %+v", pipeline[1])
	}
	lookupPipeline, ok := lookupStage["pipeline"].(bson.A)
	if !ok || len(lookupPipeline) == 0 {
		t.Fatalf("expected lookup pipeline, got %+v", lookupStage["pipeline"])
	}
	orderMatchStage, ok := lookupPipeline[0].(bson.M)
	if !ok {
		t.Fatalf("expected lookup match stage, got %+v", lookupPipeline[0])
	}
	match, ok := orderMatchStage["$match"].(bson.M)
	if !ok {
		t.Fatalf("expected lookup $match, got %+v", orderMatchStage)
	}
	expr, ok := match["$expr"].(bson.M)
	if !ok {
		t.Fatalf("expected $expr in lookup match, got %+v", match)
	}
	andItems, ok := expr["$and"].(bson.A)
	if !ok {
		t.Fatalf("expected $and conditions, got %+v", expr["$and"])
	}

	foundStatus := false
	foundBranch := false
	for _, item := range andItems {
		condition, ok := item.(bson.M)
		if !ok {
			continue
		}
		eqItems, ok := condition["$eq"].(bson.A)
		if !ok || len(eqItems) != 2 {
			continue
		}
		if eqItems[0] == "$status" && eqItems[1] == constant.ACTIVE {
			foundStatus = true
		}
		if eqItems[0] == "$branchId" && eqItems[1] == branchID {
			foundBranch = true
		}
	}

	if !foundStatus {
		t.Fatal("expected ACTIVE order filter in ABC lookup")
	}
	if !foundBranch {
		t.Fatal("expected branch filter in ABC lookup")
	}
}
