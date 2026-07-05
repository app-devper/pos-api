package repositories

import (
	"testing"

	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildGetProductAllPipelineScopesStocksByBranch(t *testing.T) {
	branchID := primitive.NewObjectID()

	pipeline, err := buildGetProductAllPipeline(request.GetProduct{BranchId: branchID.Hex()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lookupStage, ok := pipeline[3]["$lookup"].(bson.M)
	if !ok {
		t.Fatalf("expected stock lookup stage, got %+v", pipeline[3])
	}
	if _, ok := lookupStage["pipeline"].(bson.A); !ok {
		t.Fatalf("expected branch-scoped lookup pipeline, got %+v", lookupStage)
	}

	lookupPipeline := lookupStage["pipeline"].(bson.A)
	matchStage, ok := lookupPipeline[0].(bson.M)
	if !ok {
		t.Fatalf("expected lookup match stage, got %+v", lookupPipeline[0])
	}
	match, ok := matchStage["$match"].(bson.M)
	if !ok {
		t.Fatalf("expected nested $match, got %+v", matchStage)
	}
	expr, ok := match["$expr"].(bson.M)
	if !ok {
		t.Fatalf("expected $expr, got %+v", match)
	}
	andItems, ok := expr["$and"].(bson.A)
	if !ok {
		t.Fatalf("expected $and conditions, got %+v", expr["$and"])
	}

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
		if eqItems[0] == "$branchId" && eqItems[1] == branchID {
			foundBranch = true
		}
	}
	if !foundBranch {
		t.Fatal("expected branch filter in product stock lookup")
	}
}

func TestBuildGetProductAllPipelineRejectsInvalidBranchId(t *testing.T) {
	if _, err := buildGetProductAllPipeline(request.GetProduct{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}
