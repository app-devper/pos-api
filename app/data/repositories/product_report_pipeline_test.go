package repositories

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func TestBuildStockReportPipelineSeparatesByUnit(t *testing.T) {
	pipeline, err := buildStockReportPipeline("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipeline) < 6 {
		t.Fatalf("expected pipeline stages, got %d", len(pipeline))
	}

	groupStage, ok := pipeline[1]["$group"].(bson.M)
	if !ok {
		t.Fatalf("expected $group stage at index 1")
	}
	groupID, ok := groupStage["_id"].(bson.M)
	if !ok {
		t.Fatalf("expected compound _id in $group")
	}
	if groupID["unitId"] != "$unitId" {
		t.Fatalf("expected group to include unitId, got %+v", groupID)
	}

	unitLookup, ok := pipeline[4]["$lookup"].(bson.M)
	if !ok || unitLookup["from"] != "product_units" {
		t.Fatalf("expected lookup to product_units, got %+v", pipeline[4])
	}
}

func TestBuildLowStockPipelineSeparatesByUnit(t *testing.T) {
	pipeline, err := buildLowStockProductsPipeline(10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipeline) < 7 {
		t.Fatalf("expected pipeline stages, got %d", len(pipeline))
	}

	groupStage, ok := pipeline[1]["$group"].(bson.M)
	if !ok {
		t.Fatalf("expected $group stage at index 1")
	}
	groupID, ok := groupStage["_id"].(bson.M)
	if !ok {
		t.Fatalf("expected compound _id in $group")
	}
	if groupID["unitId"] != "$unitId" {
		t.Fatalf("expected group to include unitId, got %+v", groupID)
	}

	unitLookup, ok := pipeline[5]["$lookup"].(bson.M)
	if !ok || unitLookup["from"] != "product_units" {
		t.Fatalf("expected lookup to product_units, got %+v", pipeline[5])
	}
}

func TestBuildDeadStockPipelineRequiresActiveOrdersForLastSale(t *testing.T) {
	branchID := primitive.NewObjectID()
	pipeline, err := buildDeadStockProductsPipeline(time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC), branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipeline) < 6 {
		t.Fatalf("expected pipeline stages, got %d", len(pipeline))
	}

	var lookupStage bson.M
	foundOrderItemsLookup := false
	for _, stage := range pipeline {
		lookup, ok := stage["$lookup"].(bson.M)
		if ok && lookup["from"] == "order_items" {
			lookupStage = lookup
			foundOrderItemsLookup = true
			break
		}
	}
	if !foundOrderItemsLookup {
		t.Fatalf("expected order_items lookup in pipeline, got %+v", pipeline)
	}
	lookupPipeline, ok := lookupStage["pipeline"].(bson.A)
	if !ok || len(lookupPipeline) < 2 {
		t.Fatalf("expected nested pipeline in order_items lookup, got %+v", lookupStage["pipeline"])
	}
	orderLookupStage, ok := lookupPipeline[1].(bson.M)
	if !ok {
		t.Fatalf("expected nested order lookup stage, got %+v", lookupPipeline[1])
	}
	orderLookup, ok := orderLookupStage["$lookup"].(bson.M)
	if !ok || orderLookup["from"] != "orders" {
		t.Fatalf("expected nested lookup to orders, got %+v", orderLookupStage)
	}
	orderPipeline, ok := orderLookup["pipeline"].(bson.A)
	if !ok || len(orderPipeline) == 0 {
		t.Fatalf("expected nested orders pipeline, got %+v", orderLookup["pipeline"])
	}
	orderMatchStage, ok := orderPipeline[0].(bson.M)
	if !ok {
		t.Fatalf("expected first orders pipeline stage to be $match, got %+v", orderPipeline[0])
	}
	matchStage, ok := orderMatchStage["$match"].(bson.M)
	if !ok {
		t.Fatalf("expected orders $match stage, got %+v", orderMatchStage)
	}
	expr, ok := matchStage["$expr"].(bson.M)
	if !ok {
		t.Fatalf("expected $expr inside orders match, got %+v", matchStage)
	}
	andExpr, ok := expr["$and"].(bson.A)
	if !ok {
		t.Fatalf("expected $and in orders $expr, got %+v", expr)
	}
	foundActiveStatus := false
	for _, item := range andExpr {
		cond, ok := item.(bson.M)
		if !ok {
			continue
		}
		eqExpr, ok := cond["$eq"].(bson.A)
		if !ok || len(eqExpr) != 2 {
			continue
		}
		if eqExpr[0] == "$status" && eqExpr[1] == constant.ACTIVE {
			foundActiveStatus = true
			break
		}
	}
	if !foundActiveStatus {
		t.Fatalf("expected dead stock pipeline to require ACTIVE order status in nested lookup, got %+v", andExpr)
	}
}

func TestBuildExpiringProductStocksPipelineIncludesBranchFilter(t *testing.T) {
	branchID := primitive.NewObjectID()
	startDate := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	pipeline, err := buildExpiringProductStocksPipeline(request.GetProductLotsExpireRange{
		StartDate: request.NewFlexibleTime(startDate),
		EndDate:   request.NewFlexibleTime(endDate),
	}, branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipeline) == 0 {
		t.Fatalf("expected pipeline stages")
	}

	matchStage, ok := pipeline[0]["$match"].(bson.M)
	if !ok {
		t.Fatalf("expected first stage to be $match, got %+v", pipeline[0])
	}
	if matchStage["branchId"] != branchID {
		t.Fatalf("expected branchId %s in match stage, got %+v", branchID.Hex(), matchStage["branchId"])
	}
	quantityFilter, ok := matchStage["quantity"].(bson.M)
	if !ok {
		t.Fatalf("expected quantity filter, got %+v", matchStage["quantity"])
	}
	if quantityFilter["$gt"] != 0 {
		t.Fatalf("expected positive quantity filter, got %+v", quantityFilter)
	}
	expireDateFilter, ok := matchStage["expireDate"].(bson.M)
	if !ok {
		t.Fatalf("expected expireDate filter, got %+v", matchStage["expireDate"])
	}
	if expireDateFilter["$gte"] != startDate || expireDateFilter["$lt"] != endDate {
		t.Fatalf("unexpected expireDate filter: %+v", expireDateFilter)
	}
}
