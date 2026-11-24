package legacy

import (
	"context"
	"fmt"
	"sort"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Finding kinds, in the order findings reports them.
const (
	KindNotReplicaSet         = "NOT_REPLICA_SET"
	KindNegativeStock         = "NEGATIVE_STOCK"
	KindDuplicateSerial       = "DUPLICATE_SERIAL"
	KindDuplicateSequence     = "DUPLICATE_SEQUENCE"
	KindOrphanStock           = "ORPHAN_STOCK"
	KindNoMainUnit            = "NO_MAIN_UNIT"
	KindDiscountPerLine       = "DISCOUNT_PER_LINE"
	KindDiscountNotSubtracted = "DISCOUNT_NOT_SUBTRACTED"
	KindDiscountUnknown       = "DISCOUNT_UNKNOWN"
	KindUnknownPaymentType    = "UNKNOWN_PAYMENT_TYPE"
)

const (
	maxFindingIds   = 20
	amountTolerance = 0.005
)

// check is one read-only look at the database.
type check func(ctx context.Context, db *mongo.Database) ([]Finding, error)

// checks run in this order, so two runs report in the same order.
var checks = []check{
	checkReplicaSet,
	checkNegativeStock,
	checkDuplicateSerial,
	checkDuplicateSequence,
	checkOrphanStock,
	checkNoMainUnit,
	checkDiscountConvention,
	checkPaymentTypes,
}

// findings reads the database and returns what a person must decide before
// the shop moves to the current API. It never writes.
func findings(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	all := []Finding{}
	for _, run := range checks {
		found, err := run(ctx, db)
		if err != nil {
			return nil, err
		}
		all = append(all, found...)
	}
	return all, nil
}

// idFinding runs pipeline on coll; every document it returns is one offender.
// It returns no finding when there are none.
func idFinding(ctx context.Context, db *mongo.Database, coll string, pipeline mongo.Pipeline, kind, detail string) ([]Finding, error) {
	cursor, err := db.Collection(coll).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("check %s on %s: %w", kind, coll, err)
	}
	defer cursor.Close(ctx)
	var rows []struct {
		Id primitive.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("read %s on %s: %w", kind, coll, err)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id.Hex())
	}
	return findingOf(kind, detail, ids), nil
}

// findingOf is one finding for the hex ids, sorted and capped; none when ids
// is empty. ids is not changed.
func findingOf(kind, detail string, ids []string) []Finding {
	if len(ids) == 0 {
		return nil
	}
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	total := int64(len(ids))
	if len(ids) > maxFindingIds {
		ids = ids[:maxFindingIds]
	}
	return []Finding{{Kind: kind, Detail: detail, Ids: ids, Total: total}}
}

func checkReplicaSet(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	var reply bson.M
	if err := db.RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&reply); err != nil {
		return nil, fmt.Errorf("check %s: %w", KindNotReplicaSet, err)
	}
	if isReplicaSet(reply) {
		return nil, nil
	}
	return []Finding{{
		Kind:   KindNotReplicaSet,
		Detail: "MongoDB is not a replica set; every Stock write of the current API is a transaction",
		Total:  1,
	}}, nil
}

// mongosMsg is what a mongos router answers in the msg field of hello.
const mongosMsg = "isdbgrid"

// isReplicaSet reports whether a hello reply comes from a server that runs
// transactions: a replica set member, or a mongos in front of a sharded cluster.
func isReplicaSet(reply bson.M) bool {
	if msg, ok := reply["msg"].(string); ok && msg == mongosMsg {
		return true
	}
	name, ok := reply["setName"].(string)
	return ok && name != ""
}

func checkNegativeStock(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"quantity": bson.M{"$lt": 0}}}},
		{{Key: "$project", Value: bson.M{"_id": 1}}},
	}
	return idFinding(ctx, db, "product_stocks", pipeline, KindNegativeStock,
		"Stock below zero; the current Stock ledger refuses a move that leaves one")
}

// duplicatesOf lists every document sharing the value of field with another.
func duplicatesOf(field string) mongo.Pipeline {
	return mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$" + field, "ids": bson.M{"$push": "$_id"}, "n": bson.M{"$sum": 1}}}},
		{{Key: "$match", Value: bson.M{"n": bson.M{"$gt": 1}}}},
		{{Key: "$unwind", Value: "$ids"}},
		{{Key: "$project", Value: bson.M{"_id": "$ids"}}},
	}
}

// checkDuplicateSerial counts soft-deleted Products too: the unique index on
// products.serialNumber is not partial.
func checkDuplicateSerial(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	return idFinding(ctx, db, "products", duplicatesOf("serialNumber"), KindDuplicateSerial,
		"Products share a serialNumber (empty counts); the unique index on products.serialNumber cannot be built")
}

func checkDuplicateSequence(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	return idFinding(ctx, db, "sequences", duplicatesOf("field"), KindDuplicateSequence,
		"sequences share a field; the unique index on sequences.field cannot be built")
}

func checkOrphanStock(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{"from": "products", "localField": "productId", "foreignField": "_id", "as": "p"}}},
		{{Key: "$lookup", Value: bson.M{"from": "product_units", "localField": "unitId", "foreignField": "_id", "as": "u"}}},
		{{Key: "$match", Value: bson.M{"$or": bson.A{bson.M{"p": bson.M{"$size": 0}}, bson.M{"u": bson.M{"$size": 0}}}}}},
		{{Key: "$project", Value: bson.M{"_id": 1}}},
	}
	return idFinding(ctx, db, "product_stocks", pipeline, KindOrphanStock,
		"Stock whose productId or unitId matches no Product or Unit")
}

// checkNoMainUnit looks for the row catalogue.MainUnit matches: a unit of the
// Product with the Product's own unit name. Soft-deleted Products are skipped,
// as nothing is received into them.
func checkNoMainUnit(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	sameUnit := bson.M{"$expr": bson.M{"$and": bson.A{
		bson.M{"$eq": bson.A{"$productId", "$$pid"}},
		bson.M{"$eq": bson.A{"$unit", "$$name"}},
	}}}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deletedDate": bson.M{"$exists": false}}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "product_units", "let": bson.M{"pid": "$_id", "name": "$unit"},
			"pipeline": mongo.Pipeline{{{Key: "$match", Value: sameUnit}}, {{Key: "$limit", Value: 1}}},
			"as":       "main",
		}}},
		{{Key: "$match", Value: bson.M{"main": bson.M{"$size": 0}}}},
		{{Key: "$project", Value: bson.M{"_id": 1}}},
	}
	return idFinding(ctx, db, "products", pipeline, KindNoMainUnit,
		"Products whose unit name has no product_units row; a Receive is entered in the main unit")
}

func checkPaymentTypes(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	known := bson.A{constant.PaymentTypeCash, constant.PaymentTypeCredit, constant.PaymentTypePromptPay, constant.PaymentTypeTransfer}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"type": bson.M{"$nin": known}}}},
		{{Key: "$project", Value: bson.M{"_id": 1}}},
	}
	return idFinding(ctx, db, "payments", pipeline, KindUnknownPaymentType,
		"payments.type outside CASH, CREDIT, PROMPTPAY, TRANSFER")
}
