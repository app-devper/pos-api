package ledger

import (
	"context"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ImportReceive moves a Receive's lines into new Stocks of each Product's main
// Unit, settles waiting Lines from them, and marks the Receive IMPORTED — all
// in one transaction. A Receive imports once; a cancelled one never.
func (l *Ledger) ImportReceive(ctx context.Context, receiveID, branchID, by string) (*entities.Receive, error) {
	id, err := primitive.ObjectIDFromHex(receiveID)
	if err != nil {
		return nil, reject("invalid receive id")
	}
	branch, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return nil, reject("invalid branch id")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		return b.importReceive(id, branch, by)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.Receive), nil
}

func (b *book) importReceive(id, branch primitive.ObjectID, by string) (*entities.Receive, error) {
	var receive entities.Receive
	err := b.col("receives").FindOne(b.ctx, bson.M{"_id": id, "branchId": branch}).Decode(&receive)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	switch receive.Status {
	case constant.IMPORTED:
		return nil, reject("receive already imported")
	case constant.CANCELLED:
		return nil, reject("ใบรับสินค้านี้ถูกยกเลิกแล้ว นำเข้าสต็อกไม่ได้")
	}
	cursor, err := b.col("receive_items").Find(b.ctx, bson.M{"receiveId": id})
	if err != nil {
		return nil, err
	}
	items := []entities.ReceiveItem{}
	if err := cursor.All(b.ctx, &items); err != nil {
		return nil, err
	}

	now := time.Now()
	var totalCost float64
	for _, item := range items {
		totalCost += item.CostPrice * float64(item.Quantity)
		if item.Quantity <= 0 {
			continue
		}
		var product entities.Product
		if err := missing(b.col("products").FindOne(b.ctx, bson.M{"_id": item.ProductId}).Decode(&product), "product %s of a receive line", item.ProductId.Hex()); err != nil {
			return nil, err
		}
		// A Receive is entered in the Product's main Unit.
		var unit entities.ProductUnit
		if err := missing(b.col("product_units").FindOne(b.ctx, bson.M{"productId": item.ProductId, "unit": product.Unit}).Decode(&unit), "main Unit of product %s", item.ProductId.Hex()); err != nil {
			return nil, err
		}
		sequence, err := b.nextSequence(item.ProductId, unit.Id, branch)
		if err != nil {
			return nil, err
		}
		st := entities.ProductStock{Id: primitive.NewObjectID(), BranchId: branch, ProductId: item.ProductId, UnitId: unit.Id,
			ReceiveCode: receive.Code, Sequence: sequence, LotNumber: item.LotNumber, CostPrice: item.CostPrice,
			Import: item.Quantity, Quantity: item.Quantity, ExpireDate: item.ExpireDate, ImportDate: now}
		history := request.ProductStock{ProductId: item.ProductId.Hex(), UnitId: unit.Id.Hex(), ReceiveCode: receive.Code,
			Quantity: item.Quantity, CostPrice: item.CostPrice, LotNumber: item.LotNumber, ExpireDate: request.NewFlexibleTime(item.ExpireDate),
			ImportDate: request.NewFlexibleTime(now), UpdatedBy: by, BranchId: branch.Hex()}
		if _, err := b.open(st, func(int) request.ProductHistory {
			return request.AddProductStockHistory(item.ProductId.Hex(), product.Unit, history, 0)
		}); err != nil {
			return nil, err
		}
	}

	after := options.After
	var result entities.Receive
	err = b.col("receives").FindOneAndUpdate(b.ctx, bson.M{"_id": id, "status": receive.Status}, bson.M{"$set": bson.M{
		"totalCost": totalCost, "status": constant.IMPORTED, "updatedBy": by, "updatedDate": now,
	}}, &options.FindOneAndUpdateOptions{ReturnDocument: &after}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	result.Items = items
	return &result, nil
}
