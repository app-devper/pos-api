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

// receiveWrites counts the legacy Receives updated, and among them those whose
// status became IMPORTED (it was something else) and those given their
// receive_items rows as items.
type receiveWrites struct {
	updated, statused, filled int64
}

// withoutImportedStatus matches a legacy Receive whose status is not IMPORTED yet.
var withoutImportedStatus = bson.M{branchIdField: bson.M{"$exists": false}, "status": bson.M{"$ne": constant.IMPORTED}}

// changes are the Change lines for these counts.
func (w receiveWrites) changes() []Change {
	return []Change{
		{Collection: receivesCollection, Field: branchIdField, Count: w.updated},
		{Collection: receivesCollection, Field: "status", Count: w.statused},
		{Collection: receivesCollection, Field: "items", Count: w.filled},
	}
}

// migrateReceives turns the 2024 Receives (no branchId) into the current shape:
// the HQ Branch, status IMPORTED (their Stock was created at once, so they must
// never be imported again) and the items they kept in receive_items (a Receive
// with no receive_items rows keeps its items field as it is). A Receive the
// current API wrote has a branchId and is never touched; receive_items is left
// as it is. With apply set the counts are the Receives the database modified;
// otherwise those that would change. On an error during apply the counts are
// the Receives modified before it.
func migrateReceives(ctx context.Context, db *mongo.Database, hq primitive.ObjectID, apply bool) ([]Change, error) {
	if apply {
		written, err := writeReceives(ctx, db, hq)
		return written.changes(), err
	}
	legacy, err := db.Collection(receivesCollection).CountDocuments(ctx, withoutBranchId)
	if err != nil {
		return nil, fmt.Errorf("count receives without branchId: %w", err)
	}
	needStatus, err := db.Collection(receivesCollection).CountDocuments(ctx, withoutImportedStatus)
	if err != nil {
		return nil, fmt.Errorf("count receives without IMPORTED status: %w", err)
	}
	needItems, err := countReceivesToFill(ctx, db)
	if err != nil {
		return nil, err
	}
	return receiveWrites{updated: legacy, statused: needStatus, filled: needItems}.changes(), nil
}

// legacyReceive is what writeReceives reads of a 2024 Receive.
type legacyReceive struct {
	Id     primitive.ObjectID `bson:"_id"`
	Status any                `bson:"status"`
	Items  []bson.M           `bson:"items"`
}

// writeReceives visits each legacy Receive once; the branchId is set last, so
// a run that stops half way finds the same Receives again.
func writeReceives(ctx context.Context, db *mongo.Database, hq primitive.ObjectID) (receiveWrites, error) {
	cursor, err := db.Collection(receivesCollection).Find(ctx, withoutBranchId, options.Find().SetProjection(bson.M{"_id": 1, "status": 1, "items": 1}).SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return receiveWrites{}, fmt.Errorf("find legacy receives: %w", err)
	}
	var legacy []legacyReceive
	if err := cursor.All(ctx, &legacy); err != nil {
		return receiveWrites{}, fmt.Errorf("read legacy receives: %w", err)
	}
	written := receiveWrites{}
	for _, receive := range legacy {
		modified, filled, err := writeReceive(ctx, db, receive, hq)
		if err != nil {
			return written, err
		}
		if modified {
			written = receiveWrites{
				updated:  written.updated + 1,
				statused: written.statused + boolCount(receive.Status != constant.IMPORTED),
				filled:   written.filled + boolCount(filled),
			}
		}
	}
	return written, nil
}

// writeReceive migrates one legacy Receive and says whether the database
// modified it and whether it was given items.
func writeReceive(ctx context.Context, db *mongo.Database, receive legacyReceive, hq primitive.ObjectID) (bool, bool, error) {
	set := bson.M{"status": constant.IMPORTED, branchIdField: hq}
	if len(receive.Items) == 0 {
		items, err := legacyReceiveItems(ctx, db, receive.Id)
		if err != nil {
			return false, false, err
		}
		if len(items) > 0 {
			set["items"] = items
		}
	}
	filter := bson.M{"_id": receive.Id, branchIdField: bson.M{"$exists": false}}
	res, err := db.Collection(receivesCollection).UpdateOne(ctx, filter, bson.M{"$set": set})
	if err != nil {
		return false, false, fmt.Errorf("migrate receive %s: %w", receive.Id.Hex(), err)
	}
	_, filled := set["items"]
	return res.ModifiedCount == 1, filled, nil
}

// boolCount is 1 for true, 0 for false.
func boolCount(b bool) int64 {
	if b {
		return 1
	}
	return 0
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
