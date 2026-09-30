package repositories

import (
	"context"
	"fmt"
	"time"

	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// stockRecorder keeps Stock and its documents in the same session. Its helpers
// never start their own context: a failed history/document write rolls back the
// Stock change, and a concurrent writer causes the whole command to retry.
type stockRecorder struct {
	client *mongo.Client
	db     *mongo.Database
}

func newStockRecorder(resource *db.Resource) *stockRecorder {
	return &stockRecorder{client: resource.Client, db: resource.PosDb}
}

func (r *stockRecorder) transaction(command func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	session, err := r.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	return session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return command(sc)
	}, options.Transaction().SetReadConcern(readconcern.Snapshot()).SetWriteConcern(writeconcern.Majority()))
}

func (r *stockRecorder) nextCode(ctx context.Context, field, prefix string) (string, error) {
	sequence, err := (&sequenceEntity{sequenceRepo: r.db.Collection("sequences")}).nextSequenceWithContext(ctx, field)
	if err != nil {
		return "", err
	}
	return prefix + sequence.GenerateCode(), nil
}

func (r *stockRecorder) stock(ctx context.Context, id, product, branch primitive.ObjectID) (*entities.ProductStock, error) {
	var stock entities.ProductStock
	if err := r.db.Collection("product_stocks").FindOne(ctx, bson.M{"_id": id, "productId": product, "branchId": branch}).Decode(&stock); err != nil {
		return nil, fmt.Errorf("stock %s not found for this Product and branch: %w", id.Hex(), err)
	}
	return &stock, nil
}

func (r *stockRecorder) history(ctx context.Context, stock *entities.ProductStock, h request.ProductHistory) error {
	// Unit ownership and balance are checked using the same transaction snapshot.
	var unit entities.ProductUnit
	if err := r.db.Collection("product_units").FindOne(ctx, bson.M{"_id": stock.UnitId, "productId": stock.ProductId}).Decode(&unit); err != nil {
		return fmt.Errorf("stock unit not found: %w", err)
	}
	cursor, err := r.db.Collection("product_stocks").Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"productId": stock.ProductId, "unitId": stock.UnitId, "branchId": stock.BranchId}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var balances []struct {
		Balance int `bson:"balance"`
	}
	if err := cursor.All(ctx, &balances); err != nil {
		return err
	}
	balance := 0
	if len(balances) > 0 {
		balance = balances[0].Balance
	}
	if h.Unit == "" {
		h.Description += unit.Unit
	}
	_, err = r.db.Collection("product_histories").InsertOne(ctx, entities.ProductHistory{
		Id: primitive.NewObjectID(), BranchId: stock.BranchId, ProductId: stock.ProductId,
		Type: h.Type, Description: h.Description, Unit: unit.Unit, Import: h.Import, Quantity: h.Quantity,
		CostPrice: h.CostPrice, Price: h.Price, Balance: balance, CreatedBy: h.CreatedBy, CreatedDate: time.Now(),
	})
	return err
}
