package legacy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	branchesCollection = "branches"
	hqCode             = "HQ"
	hqName             = "สำนักงานใหญ่"
	systemUser         = "system"
	branchIdField      = "branchId"
)

// branchedCollections hold documents the 2024 build wrote without a branchId.
var branchedCollections = []string{
	"product_stocks",
	"orders",
	"order_items",
	"payments",
	"product_histories",
	"product_lots",
}

var withoutBranchId = bson.M{branchIdField: bson.M{"$exists": false}}

// Migrate gives every document without a branchId the HQ Branch. With apply
// false it only reports what it would change and writes nothing. On an error
// it returns the report built so far with the error: the findings, the HQ
// Branch and the Changes completed before the failure.
func Migrate(ctx context.Context, db *mongo.Database, apply bool) (Report, error) {
	found, err := findings(ctx, db)
	if err != nil {
		return Report{}, err
	}
	report := Report{Reached: StageFindings, Findings: found}
	hq, err := findHQ(ctx, db)
	if err != nil {
		return report, err
	}
	report.HQ = hq
	report.Reached = StageHQ
	if !apply {
		return migrateCollections(ctx, db, report, apply)
	}

	report.Reached = StageWriting
	if hq.IsZero() {
		hq, err = insertHQ(ctx, db)
		if err != nil {
			return report, err
		}
		report.HQ = hq
		report.HQCreated = true
	}
	return migrateCollections(ctx, db, report, apply)
}

// migrateCollections adds to a copy of report one Change per field it sets,
// stopping at the first error with the Changes done so far.
func migrateCollections(ctx context.Context, db *mongo.Database, report Report, apply bool) (Report, error) {
	changes := append([]Change(nil), report.Changes...)
	for _, name := range branchedCollections {
		change, err := backfillBranchId(ctx, db.Collection(name), report.HQ, apply)
		if err != nil {
			return report.withChanges(changes), err
		}
		changes = append(changes, change)
	}
	receives, err := migrateReceives(ctx, db, report.HQ, apply)
	changes = append(changes, receives...)
	return report.withChanges(changes), err
}

// findHQ returns the HQ Branch id, or the zero id when there is none.
func findHQ(ctx context.Context, db *mongo.Database) (primitive.ObjectID, error) {
	var branch entities.Branch
	err := db.Collection(branchesCollection).FindOne(ctx, bson.M{"code": hqCode}).Decode(&branch)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return primitive.NilObjectID, nil
	}
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("find HQ branch: %w", err)
	}
	return branch.Id, nil
}

// insertHQ writes the HQ Branch exactly as initDefaultBranch does.
func insertHQ(ctx context.Context, db *mongo.Database) (primitive.ObjectID, error) {
	now := time.Now()
	branch := entities.Branch{
		Id:          primitive.NewObjectID(),
		Code:        hqCode,
		Name:        hqName,
		Status:      constant.ACTIVE,
		CreatedBy:   systemUser,
		CreatedDate: now,
		UpdatedBy:   systemUser,
		UpdatedDate: now,
	}
	if _, err := db.Collection(branchesCollection).InsertOne(ctx, branch); err != nil {
		return primitive.NilObjectID, fmt.Errorf("insert HQ branch: %w", err)
	}
	return branch.Id, nil
}

// backfillBranchId gives hq to the documents without a branchId. With apply
// set, Count is the number the database modified; otherwise it is the number
// that would change.
func backfillBranchId(ctx context.Context, coll *mongo.Collection, hq primitive.ObjectID, apply bool) (Change, error) {
	change := Change{Collection: coll.Name(), Field: branchIdField}
	if !apply {
		count, err := coll.CountDocuments(ctx, withoutBranchId)
		if err != nil {
			return Change{}, fmt.Errorf("count %s without branchId: %w", coll.Name(), err)
		}
		change.Count = count
		return change, nil
	}
	res, err := coll.UpdateMany(ctx, withoutBranchId, bson.M{"$set": bson.M{branchIdField: hq}})
	if err != nil {
		return Change{}, fmt.Errorf("set branchId on %s: %w", coll.Name(), err)
	}
	change.Count = res.ModifiedCount
	return change, nil
}
