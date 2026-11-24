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

// withoutItems matches a Receive that carries no items of its own.
var withoutItems = bson.M{"$or": bson.A{
	bson.M{"items": bson.M{"$exists": false}},
	bson.M{"items": bson.A{}},
	bson.M{"items": nil},
}}

// migrateReceives turns the 2024 Receives (no branchId) into the current shape:
// the HQ Branch, status IMPORTED (their Stock was created at once, so they must
// never be imported again) and the items they kept in receive_items. A Receive
// the current API wrote has a branchId and is never touched; receive_items is
// left as it is. Counts are the documents found before writing.
func migrateReceives(ctx context.Context, db *mongo.Database, hq primitive.ObjectID, apply bool) ([]Change, error) {
	receives := db.Collection(receivesCollection)
	legacy, err := receives.CountDocuments(ctx, withoutBranchId)
	if err != nil {
		return nil, fmt.Errorf("count receives without branchId: %w", err)
	}
	needItems, err := receives.CountDocuments(ctx, bson.M{"$and": bson.A{withoutBranchId, withoutItems}})
	if err != nil {
		return nil, fmt.Errorf("count receives without items: %w", err)
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
			set["items"] = items
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
