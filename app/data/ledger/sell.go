package ledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/sale"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrSaleConflict: the Sale's id was already recorded for a different Sale.
var ErrSaleConflict = errors.New("sale id already recorded for a different sale")

// Sold is the Order a Sale was recorded as, its Lines, and the current state
// of the Stocks it drew from.
type Sold struct {
	Order  *entities.Order
	Lines  []entities.OrderItem
	Stocks []entities.ProductStock
}

// Sell records a Sale as an Order (ADR-0001). The server prices each Line
// and draws its Stock as app/domain/sale decides; what no Stock covers goes
// to Oversell (on the Line) or Sold first (ADR-0003). The Order code is taken
// inside the transaction, so a refused Sale takes none. A Sale whose id is
// already recorded returns that Order when it is the same Sale — the same
// Lines for the same Customer, however it was paid — and ErrSaleConflict
// otherwise.
func (l *Ledger) Sell(ctx context.Context, form request.Sale) (*Sold, error) {
	branch, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, reject("invalid branch id")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		if found, err := b.findSale(form); found != nil || err != nil {
			return found, err
		}
		return b.sell(form, branch)
	})
	if mongo.IsDuplicateKeyError(err) {
		// Another request recorded this Sale's id first.
		return l.findSale(ctx, form)
	}
	if err != nil {
		return nil, err
	}
	return result.(*Sold), nil
}

func (l *Ledger) findSale(ctx context.Context, form request.Sale) (*Sold, error) {
	sold, err := (&book{l: l, ctx: ctx}).findSale(form)
	if err == nil && sold == nil {
		return nil, ErrConflict
	}
	return sold, err
}

func (b *book) findSale(form request.Sale) (*Sold, error) {
	var order entities.Order
	err := b.col("orders").FindOne(b.ctx, bson.M{"saleId": form.SaleId}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if order.SaleFingerprint != fingerprint(form) && order.SaleFingerprint != legacyFingerprint(form) {
		return nil, ErrSaleConflict
	}
	var lines []entities.OrderItem
	if err := find(b, "order_items", bson.M{"orderId": order.Id}, &lines); err != nil {
		return nil, err
	}
	if err := find(b, "payments", bson.M{"orderId": order.Id}, &order.Payments); err != nil {
		return nil, err
	}
	var ids []primitive.ObjectID
	for _, line := range lines {
		for _, s := range line.Stocks {
			if id, err := primitive.ObjectIDFromHex(s.StockId); err == nil {
				ids = append(ids, id)
			}
		}
	}
	stocks := []entities.ProductStock{}
	if len(ids) > 0 {
		if err := find(b, "product_stocks", bson.M{"_id": bson.M{"$in": ids}}, &stocks); err != nil {
			return nil, err
		}
	}
	return &Sold{Order: &order, Lines: lines, Stocks: stocks}, nil
}

func find[T any](b *book, collection string, filter bson.M, into *[]T) error {
	cursor, err := b.col(collection).Find(b.ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return err
	}
	return cursor.All(b.ctx, into)
}

// saleUnit is one Unit's Catalog for the Sale; its Stock quantities follow
// what earlier Lines of the Sale drew.
type saleUnit struct {
	unit    entities.ProductUnit
	catalog sale.Catalog
}

func (b *book) loadSaleUnit(product, unit, branch primitive.ObjectID) (*saleUnit, error) {
	u := &saleUnit{}
	err := b.col("product_units").FindOne(b.ctx, bson.M{"_id": unit, "productId": product}).Decode(&u.unit)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, reject("unit %s is not a unit of product %s", unit.Hex(), product.Hex())
	}
	if err != nil {
		return nil, err
	}
	u.catalog.UnitCost = u.unit.CostPrice
	if err := find(b, "product_prices", bson.M{"productId": product, "unitId": unit}, &u.catalog.Prices); err != nil {
		return nil, err
	}
	if err := find(b, "product_stocks", bson.M{"productId": product, "unitId": unit, "branchId": branch}, &u.catalog.Stocks); err != nil {
		return nil, err
	}
	return u, nil
}

func (b *book) sell(form request.Sale, branch primitive.ObjectID) (*Sold, error) {
	now := time.Now()
	orderID := primitive.NewObjectID()
	units := map[primitive.ObjectID]*saleUnit{}
	drawnStocks := map[primitive.ObjectID]*entities.ProductStock{}
	var drawnOrder []primitive.ObjectID
	lines := make([]entities.OrderItem, 0, len(form.Items))
	var money lineMoney

	for _, line := range form.Items {
		product, err := primitive.ObjectIDFromHex(line.ProductId)
		if err != nil {
			return nil, reject("invalid product id %q", line.ProductId)
		}
		unit, err := primitive.ObjectIDFromHex(line.UnitId)
		if err != nil {
			return nil, reject("invalid unit id %q", line.UnitId)
		}
		u, ok := units[unit]
		if !ok {
			if u, err = b.loadSaleUnit(product, unit, branch); err != nil {
				return nil, err
			}
			units[unit] = u
		} else if u.unit.ProductId != product {
			return nil, reject("unit %s is not a unit of product %s", unit.Hex(), product.Hex())
		}
		rung := sale.Ring(sale.Line{ProductId: line.ProductId, UnitId: line.UnitId, Quantity: line.Quantity, PriceType: line.PriceType,
			StockId: line.StockId, Discount: line.Discount, AllowOversell: line.AllowOversell}, u.catalog)

		lineID := primitive.NewObjectID()
		unitName, qty, amount, cost := u.unit.Unit, line.Quantity, rung.Amount, rung.Cost
		r := b.rowFor(lineID, &entities.ProductStock{BranchId: branch, ProductId: product, UnitId: unit}, func(int) request.ProductHistory {
			return request.AddOrderItemProductHistory(product.Hex(), unitName, request.OrderItem{Quantity: qty, Price: amount, CostPrice: cost}, 0, form.CreatedBy)
		})
		drawn := make([]entities.OrderItemStock, 0, len(rung.Parts))
		for i, part := range rung.Parts {
			if part.StockId == "" {
				if err := b.soldFirst(product, -part.Quantity); err != nil {
					return nil, err
				}
				drawn = append(drawn, entities.OrderItemStock{Quantity: part.Quantity})
				continue
			}
			take := part.Quantity
			if i == len(rung.Parts)-1 {
				take -= rung.Oversold
			}
			if take <= 0 {
				continue
			}
			id, _ := primitive.ObjectIDFromHex(part.StockId)
			st := u.stock(id)
			if st == nil {
				return nil, reject("stock %s is not a Stock of this Unit in this branch", part.StockId)
			}
			if err := b.move(st, -take, r); err != nil {
				return nil, err
			}
			if _, seen := drawnStocks[id]; !seen {
				drawnOrder = append(drawnOrder, id)
			}
			drawnStocks[id] = st
			drawn = append(drawn, entities.OrderItemStock{StockId: part.StockId, Quantity: take})
		}
		lines = append(lines, entities.OrderItem{Id: lineID, BranchId: branch, OrderId: orderID, ProductId: product, UnitId: unit,
			Status: constant.CONFIRMED, Stocks: drawn, Quantity: line.Quantity, Price: rung.Amount, CostPrice: rung.Cost,
			Discount: rung.Discount, OversoldQty: rung.Oversold, CreatedBy: form.CreatedBy, CreatedDate: now, UpdatedBy: form.CreatedBy, UpdatedDate: now})
		money.add(rung.Paid(line.Quantity), rung.Cost, rung.Discount*float64(line.Quantity))
	}
	total, totalCost, discount := money.rounded()

	var tendered float64
	for _, p := range form.Payments {
		if p.Amount < 0 {
			return nil, reject("payment amount must not be negative")
		}
		tendered += p.Amount
	}
	if roundMoney(tendered) < total {
		return nil, reject("payment %.2f is less than the total %.2f", tendered, total)
	}
	change := roundMoney(tendered - total)

	code, err := b.l.codes(b.ctx, constant.ORDER, "")
	if err != nil {
		return nil, err
	}
	order := entities.Order{Id: orderID, BranchId: branch, Code: code, CustomerCode: form.CustomerCode, CustomerName: form.CustomerName,
		PatientId: form.PatientId, PharmacistName: form.PharmacistName, LicenseNo: form.LicenseNo, PrescriberName: form.PrescriberName,
		BuyerName: form.BuyerName, BuyerIdCard: form.BuyerIdCard, Status: constant.CONFIRMED, Total: total, TotalCost: totalCost,
		Discount: discount, Type: form.Type, SaleId: form.SaleId, SaleFingerprint: fingerprint(form),
		CreatedBy: form.CreatedBy, CreatedDate: now, UpdatedBy: form.CreatedBy, UpdatedDate: now}
	if _, err := b.col("orders").InsertOne(b.ctx, order); err != nil {
		return nil, err
	}
	docs := make([]any, len(lines))
	for i := range lines {
		docs[i] = lines[i]
	}
	if _, err := b.col("order_items").InsertMany(b.ctx, docs); err != nil {
		return nil, err
	}
	payments := make([]any, len(form.Payments))
	for i, p := range form.Payments {
		payment := entities.Payment{Id: primitive.NewObjectID(), BranchId: branch, OrderId: orderID, Status: constant.ACTIVE,
			Amount: p.Amount, Total: total, Change: change, Type: p.Type,
			CreatedBy: form.CreatedBy, CreatedDate: now, UpdatedBy: form.CreatedBy, UpdatedDate: now}
		payments[i] = payment
		order.Payments = append(order.Payments, payment)
	}
	if _, err := b.col("payments").InsertMany(b.ctx, payments); err != nil {
		return nil, err
	}
	stocks := make([]entities.ProductStock, 0, len(drawnOrder))
	for _, id := range drawnOrder {
		stocks = append(stocks, *drawnStocks[id])
	}
	return &Sold{Order: &order, Lines: lines, Stocks: stocks}, nil
}

// stock is the Unit's Stock with that id, as this Sale has left it.
func (u *saleUnit) stock(id primitive.ObjectID) *entities.ProductStock {
	for i := range u.catalog.Stocks {
		if u.catalog.Stocks[i].Id == id {
			return &u.catalog.Stocks[i]
		}
	}
	return nil
}

// lineMoney sums what an Order's Lines charge, the way a Sale does: each
// Line's paid amount is rounded by app/domain/sale, the sums are rounded
// once. A cancel recomputes an Order the same way.
type lineMoney struct{ total, cost, discount float64 }

func (m *lineMoney) add(paid, cost, discount float64) {
	m.total += paid
	m.cost += cost
	m.discount += discount
}

func (m lineMoney) rounded() (total, cost, discount float64) {
	return roundMoney(m.total), roundMoney(m.cost), roundMoney(m.discount)
}

func roundMoney(v float64) float64 { return math.Round(v*100) / 100 }

// fingerprint identifies a Sale by what was sold to whom: its branch, its
// Lines, the Customer, and the regulated details of who it was for (patient,
// prescriber, pharmacist, buyer). How it was paid is not part of it, so a till
// resending a Sale with a corrected tender gets the Order already recorded,
// while a resend that corrects who it was for is a different Sale.
func fingerprint(form request.Sale) string {
	content, _ := json.Marshal(struct {
		BranchId, CustomerCode, CustomerName, PatientId, PharmacistName string
		LicenseNo, PrescriberName, BuyerName, BuyerIdCard               string
		Items                                                           []request.SaleLine
	}{form.BranchId, form.CustomerCode, form.CustomerName, form.PatientId, form.PharmacistName,
		form.LicenseNo, form.PrescriberName, form.BuyerName, form.BuyerIdCard, form.Items})
	sum := sha256.Sum256(content)
	return "v2:" + hex.EncodeToString(sum[:])
}

// legacyFingerprint is how Orders recorded before fingerprint were
// identified: the whole request, payments included. Still accepted, so a
// Sale recorded before the change is recognised when the till resends it.
func legacyFingerprint(form request.Sale) string {
	content, _ := json.Marshal(struct {
		BranchId string
		Sale     request.Sale
	}{form.BranchId, form})
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
