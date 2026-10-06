package ledger

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/stock"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// book is one event's view of Stock inside its transaction. Every quantity
// change goes through change or open, so the no-negative rule, settlement and
// history are kept in one place.
type book struct {
	l       *Ledger
	ctx     context.Context
	touched []*touch
	byStock map[primitive.ObjectID]*touch
	// notOwed are Lines this event must not settle — a Return's own Lines,
	// whose goods are coming back, not going out.
	notOwed []primitive.ObjectID
}

// touch is one Stock the event changed. Its history row is written once, at
// flush, with the net quantity the event moved and the final balance.
type touch struct {
	stock   *entities.ProductStock
	qty     int
	history func(qty int) request.ProductHistory
	settled int
}

func newBook(l *Ledger, ctx context.Context) *book {
	return &book{l: l, ctx: ctx, byStock: map[primitive.ObjectID]*touch{}}
}

func (b *book) col(name string) *mongo.Collection { return b.l.db.Collection(name) }

// stock loads a Stock of a Product in a branch.
func (b *book) stock(id, product, branch primitive.ObjectID) (*entities.ProductStock, error) {
	var st entities.ProductStock
	err := b.col("product_stocks").FindOne(b.ctx, bson.M{"_id": id, "productId": product, "branchId": branch}).Decode(&st)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("%w: stock %s for this Product and branch", ErrNotFound, id.Hex())
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// change moves a Stock's quantity by delta. The quantity never goes below
// zero, and a rise settles that Unit's waiting Lines from the Stock.
func (b *book) change(st *entities.ProductStock, delta int, history func(qty int) request.ProductHistory) error {
	if delta == 0 {
		return nil
	}
	after := st.Quantity + delta
	if after < 0 || (delta > 0 && after < st.Quantity) {
		return reject("insufficient stock or quantity overflow")
	}
	if err := b.inc(st, delta); err != nil {
		return err
	}
	t := b.track(st, history)
	t.qty += delta
	if delta > 0 {
		return b.settle(t, delta)
	}
	return nil
}

// open inserts a new Stock and settles that Unit's waiting Lines from it.
func (b *book) open(st entities.ProductStock, history func(qty int) request.ProductHistory) (*entities.ProductStock, error) {
	if st.Quantity < 0 {
		return nil, reject("stock quantity must not be negative")
	}
	if _, err := b.col("product_stocks").InsertOne(b.ctx, st); err != nil {
		return nil, err
	}
	t := b.track(&st, history)
	t.qty += st.Quantity
	if err := b.settle(t, st.Quantity); err != nil {
		return nil, err
	}
	return &st, nil
}

func (b *book) track(st *entities.ProductStock, history func(qty int) request.ProductHistory) *touch {
	if t, ok := b.byStock[st.Id]; ok {
		t.stock = st
		return t
	}
	t := &touch{stock: st, history: history}
	b.byStock[st.Id] = t
	b.touched = append(b.touched, t)
	return t
}

// inc applies delta only if the Stock still holds what this event read.
func (b *book) inc(st *entities.ProductStock, delta int) error {
	res, err := b.col("product_stocks").UpdateOne(b.ctx, bson.M{"_id": st.Id, "quantity": st.Quantity}, bson.M{"$inc": bson.M{"quantity": delta}})
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return ErrConflict
	}
	st.Quantity += delta
	return nil
}

// settle serves the oldest Lines still owed the Stock's Unit in its branch out
// of what just came in, drawing what they get out of the Stock (ADR-0002).
// Quantity the Stock already held is not reconsidered.
func (b *book) settle(t *touch, incoming int) error {
	st := t.stock
	incoming = min(incoming, st.Quantity)
	if incoming <= 0 {
		return nil
	}
	filter := bson.M{"branchId": st.BranchId, "productId": st.ProductId, "unitId": st.UnitId,
		"oversoldQty": bson.M{"$gt": 0}, "$or": confirmedLineStatuses()}
	if len(b.notOwed) > 0 {
		filter["_id"] = bson.M{"$nin": b.notOwed}
	}
	cursor, err := b.col("order_items").Find(b.ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return err
	}
	var lines []entities.OrderItem
	if err := cursor.All(b.ctx, &lines); err != nil {
		return err
	}
	debts := make([]stock.Debt, 0, len(lines))
	for _, line := range lines {
		debts = append(debts, stock.Debt{Line: line.Id, Owed: line.OversoldQty})
	}
	for _, s := range stock.Settle(incoming, debts) {
		res, err := b.col("order_items").UpdateOne(b.ctx, bson.M{"_id": s.Line, "oversoldQty": bson.M{"$gte": s.Qty}}, mongo.Pipeline{{{Key: "$set", Value: bson.M{
			"oversoldQty": bson.M{"$subtract": bson.A{"$oversoldQty", s.Qty}},
			"stocks":      bson.M{"$concatArrays": bson.A{bson.M{"$ifNull": bson.A{"$stocks", bson.A{}}}, bson.A{entities.OrderItemStock{StockId: st.Id.Hex(), Quantity: s.Qty}}}},
			"updatedDate": time.Now(),
		}}}})
		if err != nil {
			return err
		}
		if res.MatchedCount != 1 {
			return ErrConflict
		}
		if err := b.inc(st, -s.Qty); err != nil {
			return err
		}
		t.qty -= s.Qty
		t.settled += s.Qty
	}
	return nil
}

// flush writes one history row per Stock touched, with the balance its branch,
// Product and Unit hold after the whole event.
func (b *book) flush() error {
	now := time.Now()
	for _, t := range b.touched {
		st := t.stock
		var unit entities.ProductUnit
		if err := b.col("product_units").FindOne(b.ctx, bson.M{"_id": st.UnitId, "productId": st.ProductId}).Decode(&unit); err != nil {
			return fmt.Errorf("stock unit not found: %w", err)
		}
		balance, err := b.balance(st)
		if err != nil {
			return err
		}
		h := t.history(t.qty + t.settled)
		if h.Unit == "" {
			h.Description += unit.Unit
		}
		if t.settled > 0 {
			h.Description += " (ส่งของค้างลูกค้า " + strconv.Itoa(t.settled) + " " + unit.Unit + ")"
		}
		if _, err := b.col("product_histories").InsertOne(b.ctx, entities.ProductHistory{
			Id: primitive.NewObjectID(), BranchId: st.BranchId, ProductId: st.ProductId,
			Type: h.Type, Description: h.Description, Unit: unit.Unit, Import: h.Import, Quantity: h.Quantity,
			CostPrice: h.CostPrice, Price: h.Price, Balance: balance, CreatedBy: h.CreatedBy, CreatedDate: now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (b *book) balance(st *entities.ProductStock) (int, error) {
	cursor, err := b.col("product_stocks").Aggregate(b.ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"productId": st.ProductId, "unitId": st.UnitId, "branchId": st.BranchId}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}}},
	})
	if err != nil {
		return 0, err
	}
	var rows []struct {
		Balance int `bson:"balance"`
	}
	if err := cursor.All(b.ctx, &rows); err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Balance, nil
}

// nextSequence is the selling order a new Stock of a Unit takes in a branch.
func (b *book) nextSequence(product, unit, branch primitive.ObjectID) (int, error) {
	var last entities.ProductStock
	err := b.col("product_stocks").FindOne(b.ctx, bson.M{"productId": product, "unitId": unit, "branchId": branch},
		options.FindOne().SetSort(bson.D{{Key: "sequence", Value: -1}})).Decode(&last)
	if err == mongo.ErrNoDocuments {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.Sequence + 1, nil
}

// confirmedLineStatuses matches a Line that still stands: no status (older
// Lines) or one of constant.ConfirmedOrderStatuses.
func confirmedLineStatuses() []bson.M {
	clauses := []bson.M{{"status": bson.M{"$exists": false}}, {"status": ""}}
	for _, status := range constant.ConfirmedOrderStatuses() {
		clauses = append(clauses, bson.M{"status": status})
	}
	return clauses
}
