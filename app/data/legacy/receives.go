package legacy

import (
	"context"
	"fmt"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	receivesCollection     = "receives"
	receiveItemsCollection = "receive_items"
)

// withoutItemsAt matches a Receive, found under path, that carries no items of
// its own and no branchId.
func withoutItemsAt(path string) bson.M {
	return bson.M{
		path + branchIdField: bson.M{"$exists": false},
		"$or": bson.A{
			bson.M{path + "items": bson.M{"$exists": false}},
			bson.M{path + "items": bson.A{}},
			bson.M{path + "items": nil},
		},
	}
}

// countReceivesToFill counts the legacy Receives without items that have at
// least one receive_items row. It starts from receive_items, so it needs no
// index on receiveId: one scan, then each Receive is found by its _id.
func countReceivesToFill(ctx context.Context, db *mongo.Database) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$receiveId"}}},
		{{Key: "$lookup", Value: bson.M{"from": receivesCollection, "localField": "_id", "foreignField": "_id", "as": "r"}}},
		{{Key: "$unwind", Value: "$r"}},
		{{Key: "$match", Value: withoutItemsAt("r.")}},
		{{Key: "$count", Value: "n"}},
	}
	cursor, err := db.Collection(receiveItemsCollection).Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return 0, fmt.Errorf("count receives without items: %w", err)
	}
	defer cursor.Close(ctx)
	var rows []struct {
		N int64 `bson:"n"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, fmt.Errorf("read count of receives without items: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].N, nil
}

// migrateReceives turns the 2024 Receives (no branchId) into the current shape:
// the HQ Branch, status IMPORTED (their Stock was created at once, so they must
// never be imported again) and the items they kept in receive_items (a Receive
// with no receive_items rows keeps its items field as it is). A Receive
// the current API wrote has a branchId and is never touched; receive_items is
// left as it is. Counts are the documents found before writing.
func migrateReceives(ctx context.Context, db *mongo.Database, hq primitive.ObjectID, apply bool) ([]Change, error) {
	receives := db.Collection(receivesCollection)
	legacy, err := receives.CountDocuments(ctx, withoutBranchId)
	if err != nil {
		return nil, fmt.Errorf("count receives without branchId: %w", err)
	}
	needItems, err := countReceivesToFill(ctx, db)
	if err != nil {
		return nil, err
	}
	changes := []Change{
		{Collection: receivesCollection, Field: branchIdField, Count: legacy},
		{Collection: receivesCollection, Field: "status", Count: legacy},
		{Collection: receivesCollection, Field: "items", Count: needItems},
	}
	if !apply || legacy == 0 {
		return changes, nil
	}
	if err := writeReceives(ctx, db, hq); err != nil {
		return nil, err
	}
	return changes, nil
}

// writeReceives visits each legacy Receive once; the branchId is set last, so
// a run that stops half way finds the same Receives again.
func writeReceives(ctx context.Context, db *mongo.Database, hq primitive.ObjectID) error {
	receives := db.Collection(receivesCollection)
	cursor, err := receives.Find(ctx, withoutBranchId, options.Find().SetProjection(bson.M{"_id": 1, "items": 1}))
	if err != nil {
		return fmt.Errorf("find legacy receives: %w", err)
	}
	var legacy []struct {
		Id    primitive.ObjectID `bson:"_id"`
		Items []bson.M           `bson:"items"`
	}
	if err := cursor.All(ctx, &legacy); err != nil {
		return fmt.Errorf("read legacy receives: %w", err)
	}
	for _, receive := range legacy {
		set := bson.M{"status": constant.IMPORTED, branchIdField: hq}
		if len(receive.Items) == 0 {
			items, err := legacyReceiveItems(ctx, db, receive.Id)
			if err != nil {
				return err
			}
			if len(items) > 0 {
				set["items"] = items
			}
		}
		filter := bson.M{"_id": receive.Id, branchIdField: bson.M{"$exists": false}}
		if _, err := receives.UpdateOne(ctx, filter, bson.M{"$set": set}); err != nil {
			return fmt.Errorf("migrate receive %s: %w", receive.Id.Hex(), err)
		}
	}
	return nil
}

// legacyReceiveItems reads a Receive's lines from receive_items, oldest first,
// keeping only productId, costPrice and quantity.
func legacyReceiveItems(ctx context.Context, db *mongo.Database, receiveID primitive.ObjectID) (bson.A, error) {
	cursor, err := db.Collection(receiveItemsCollection).Find(ctx, bson.M{"receiveId": receiveID},
		options.Find().SetSort(bson.D{{Key: "createdDate", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("find items of receive %s: %w", receiveID.Hex(), err)
	}
	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("read items of receive %s: %w", receiveID.Hex(), err)
	}
	items := make(bson.A, 0, len(rows))
	for _, row := range rows {
		items = append(items, bson.M{"productId": row["productId"], "costPrice": row["costPrice"], "quantity": row["quantity"]})
	}
	return items, nil
}
