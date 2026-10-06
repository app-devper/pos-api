package ledger

import (
	"context"
	"strconv"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CrossUnitDraw is a settlement that served a Line from a Stock of another
// Unit, which ADR-0002 forbids.
type CrossUnitDraw struct {
	Line     primitive.ObjectID
	Stock    primitive.ObjectID
	Quantity int
}

// RepairCrossUnitOversell finds settlements that drew a Line's debt from a
// Stock of a different Unit. With apply, each one is undone in its own
// transaction: the quantity goes back to the Stock (settling that Unit's own
// waiting Lines, as any inflow does) and the Line is owed it again. A Sale
// only ever draws its own Unit, so every such draw is a wrong settlement.
func (l *Ledger) RepairCrossUnitOversell(ctx context.Context, apply bool, by string) ([]CrossUnitDraw, error) {
	found, err := l.findCrossUnitDraws(ctx)
	if err != nil || !apply {
		return found, err
	}
	for _, d := range found {
		if _, err := l.run(ctx, func(b *book) (any, error) { return nil, b.undoDraw(d, by) }); err != nil {
			return found, err
		}
	}
	return found, nil
}

func (l *Ledger) findCrossUnitDraws(ctx context.Context) ([]CrossUnitDraw, error) {
	cursor, err := l.db.Collection("order_items").Aggregate(ctx, mongo.Pipeline{
		// A cancelled Line already gave its draws back; a Line without a Unit
		// cannot be judged.
		{{Key: "$match", Value: bson.M{"stocks.0": bson.M{"$exists": true}, "$or": confirmedLineStatuses(),
			"unitId": bson.M{"$exists": true, "$ne": primitive.NilObjectID}}}},
		{{Key: "$unwind", Value: "$stocks"}},
		{{Key: "$match", Value: bson.M{"stocks.stockid": bson.M{"$regex": "^[0-9a-f]{24}$"}}}},
		{{Key: "$addFields", Value: bson.M{"stockObjectId": bson.M{"$toObjectId": "$stocks.stockid"}}}},
		{{Key: "$lookup", Value: bson.M{"from": "product_stocks", "localField": "stockObjectId", "foreignField": "_id", "as": "drawn"}}},
		{{Key: "$unwind", Value: "$drawn"}},
		{{Key: "$match", Value: bson.M{"$expr": bson.M{"$ne": bson.A{"$drawn.unitId", "$unitId"}}}}},
		{{Key: "$project", Value: bson.M{"line": "$_id", "stock": "$drawn._id", "quantity": "$stocks.quantity"}}},
	})
	if err != nil {
		return nil, err
	}
	var rows []struct {
		Line     primitive.ObjectID `bson:"line"`
		Stock    primitive.ObjectID `bson:"stock"`
		Quantity int                `bson:"quantity"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	found := make([]CrossUnitDraw, 0, len(rows))
	for _, r := range rows {
		found = append(found, CrossUnitDraw{Line: r.Line, Stock: r.Stock, Quantity: r.Quantity})
	}
	return found, nil
}

func (b *book) undoDraw(d CrossUnitDraw, by string) error {
	var line entities.OrderItem
	if err := b.col("order_items").FindOne(b.ctx, bson.M{"_id": d.Line}).Decode(&line); err != nil {
		return err
	}
	remaining := make([]entities.OrderItemStock, 0, len(line.Stocks))
	removed := false
	for _, s := range line.Stocks {
		if !removed && s.StockId == d.Stock.Hex() && s.Quantity == d.Quantity {
			removed = true
			continue
		}
		remaining = append(remaining, s)
	}
	if !removed {
		return nil // already repaired
	}
	if _, err := b.col("order_items").UpdateOne(b.ctx, bson.M{"_id": d.Line}, bson.M{
		"$set": bson.M{"stocks": remaining, "updatedDate": time.Now()},
		"$inc": bson.M{"oversoldQty": d.Quantity},
	}); err != nil {
		return err
	}
	var st entities.ProductStock
	if err := b.col("product_stocks").FindOne(b.ctx, bson.M{"_id": d.Stock}).Decode(&st); err != nil {
		return err
	}
	return b.change(&st, d.Quantity, func(qty int) request.ProductHistory {
		return request.ProductHistory{ProductId: st.ProductId.Hex(), Type: "REPAIR_CROSS_UNIT_OVERSELL",
			Description: "คืนสต็อกที่ถูกตัดหนี้ข้ามหน่วยผิด จำนวน " + strconv.Itoa(qty) + " ", Quantity: qty, CreatedBy: by}
	})
}
