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
// false it only reports what it would change and writes nothing.
func Migrate(ctx context.Context, db *mongo.Database, apply bool) (Report, error) {
	report := Report{}
	hq, err := findHQ(ctx, db)
	if err != nil {
		return Report{}, err
	}
	report.HQ = hq

	if apply && hq.IsZero() {
		hq, err = insertHQ(ctx, db)
		if err != nil {
			return Report{}, err
		}
		report.HQ = hq
		report.HQCreated = true
	}

	for _, name := range branchedCollections {
		change, err := backfillBranchId(ctx, db.Collection(name), report.HQ, apply)
		if err != nil {
			return Report{}, err
		}
		report.Changes = append(report.Changes, change)
	}
	return report, nil
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

// backfillBranchId counts the documents without a branchId and, when apply is
// set, gives them hq. Count is the number found before writing.
func backfillBranchId(ctx context.Context, coll *mongo.Collection, hq primitive.ObjectID, apply bool) (Change, error) {
	change := Change{Collection: coll.Name(), Field: branchIdField}
	count, err := coll.CountDocuments(ctx, withoutBranchId)
	if err != nil {
		return Change{}, fmt.Errorf("count %s without branchId: %w", coll.Name(), err)
	}
	change.Count = count
	if !apply || count == 0 {
		return change, nil
	}
	if _, err := coll.UpdateMany(ctx, withoutBranchId, bson.M{"$set": bson.M{branchIdField: hq}}); err != nil {
		return Change{}, fmt.Errorf("set branchId on %s: %w", coll.Name(), err)
	}
	return change, nil
}
