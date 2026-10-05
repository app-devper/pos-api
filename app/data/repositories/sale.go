package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/sale"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrSaleConflict: the Sale's id was already recorded for a different Sale.
var ErrSaleConflict = errors.New("sale id already recorded for a different sale")

// SaleRejected is a Sale that cannot be recorded as it stands.
type SaleRejected struct{ Reason string }

func (e *SaleRejected) Error() string { return e.Reason }

func rejectSale(format string, args ...any) error {
	return &SaleRejected{Reason: fmt.Sprintf(format, args...)}
}

// RecordedSale is the Order a Sale was recorded as, with the current state
// of the Stocks it drew from.
type RecordedSale struct {
	Order  *entities.Order
	Stocks []entities.ProductStock
}

// saleFingerprint identifies a Sale's content, so a repeat of the Sale can be
// told from a different Sale reusing its id.
func saleFingerprint(form request.Sale) string {
	content, _ := json.Marshal(struct {
		BranchId string
		Sale     request.Sale
	}{form.BranchId, form})
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// FindSale returns the Order already recorded for the Sale's id, or nil if
// there is none; ErrSaleConflict if the id was used for a different Sale.
func (entity *orderEntity) FindSale(form request.Sale) (*RecordedSale, error) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.findSaleWithContext(ctx, form)
}

func (entity *orderEntity) findSaleWithContext(ctx context.Context, form request.Sale) (*RecordedSale, error) {
	var order entities.Order
	err := entity.orderRepo.FindOne(ctx, bson.M{"saleId": form.SaleId}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if order.SaleFingerprint != saleFingerprint(form) {
		return nil, ErrSaleConflict
	}
	orders := []entities.Order{order}
	if err := entity.populateOrderPaymentsWithContext(ctx, orders); err != nil {
		return nil, err
	}
	var items []entities.OrderItem
	cursor, err := entity.orderItemRepo.Find(ctx, bson.M{"orderId": order.Id})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	var stockIds []primitive.ObjectID
	for _, item := range items {
		for _, s := range item.Stocks {
			if id, err := primitive.ObjectIDFromHex(s.StockId); err == nil {
				stockIds = append(stockIds, id)
			}
		}
	}
	stocks := []entities.ProductStock{}
	if len(stockIds) > 0 {
		cursor, err = entity.productStockRepo.Find(ctx, bson.M{"_id": bson.M{"$in": stockIds}})
		if err != nil {
			return nil, err
		}
		if err := cursor.All(ctx, &stocks); err != nil {
			return nil, err
		}
	}
	return &RecordedSale{Order: &orders[0], Stocks: stocks}, nil
}

// RecordSale records a Sale as an Order in one transaction: it prices each
// Line and draws its Stock as app/domain/sale decides from the Stock as it
// stands, adds the product history, and takes payment. Nothing is written
// unless all of it is. A Sale whose id is already recorded returns that
// Order (or ErrSaleConflict); a Sale that cannot be recorded returns
// *SaleRejected.
func (entity *orderEntity) RecordSale(form request.Sale) (*RecordedSale, error) {
	logrus.Info("RecordSale")
	ctx, cancel := utils.InitContext()
	defer cancel()

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var recorded *RecordedSale
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		r, txErr := entity.recordSaleWithContext(sessCtx, form)
		recorded = r
		return nil, txErr
	})
	if mongo.IsDuplicateKeyError(err) {
		// The Sale's id is already recorded: a repeat, or a different Sale.
		if found, findErr := entity.FindSale(form); findErr != nil || found != nil {
			return found, findErr
		}
	}
	if err != nil {
		return nil, err
	}
	return recorded, nil
}

// saleUnit is one Unit's Catalog for the Sale; its Stock quantities follow
// what earlier Lines of the Sale drew.
type saleUnit struct {
	unit    entities.ProductUnit
	catalog sale.Catalog
}

func (entity *orderEntity) loadSaleUnit(ctx context.Context, productId, unitId, branchId primitive.ObjectID) (*saleUnit, error) {
	u := &saleUnit{}
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"_id": unitId, "productId": productId}).Decode(&u.unit)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, rejectSale("unit %s is not a unit of product %s", unitId.Hex(), productId.Hex())
	}
	if err != nil {
		return nil, err
	}
	u.catalog.UnitCost = u.unit.CostPrice

	cursor, err := entity.productPricesRepo.Find(ctx, bson.M{"productId": productId, "unitId": unitId},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &u.catalog.Prices); err != nil {
		return nil, err
	}
	cursor, err = entity.productStockRepo.Find(ctx, bson.M{"productId": productId, "unitId": unitId, "branchId": branchId})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &u.catalog.Stocks); err != nil {
		return nil, err
	}
	return u, nil
}

func (entity *orderEntity) recordSaleWithContext(ctx context.Context, form request.Sale) (*RecordedSale, error) {
	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	orderId := primitive.NewObjectID()
	units := map[primitive.ObjectID]*saleUnit{}
	updated := map[string]entities.ProductStock{}
	var updatedOrder []string
	items := make([]interface{}, 0, len(form.Items))
	var total, totalCost, discount float64

	for _, line := range form.Items {
		productId, err := primitive.ObjectIDFromHex(line.ProductId)
		if err != nil {
			return nil, rejectSale("invalid product id %q", line.ProductId)
		}
		unitId, err := primitive.ObjectIDFromHex(line.UnitId)
		if err != nil {
			return nil, rejectSale("invalid unit id %q", line.UnitId)
		}
		u, ok := units[unitId]
		if !ok {
			if u, err = entity.loadSaleUnit(ctx, productId, unitId, branchId); err != nil {
				return nil, err
			}
			units[unitId] = u
		} else if u.unit.ProductId != productId {
			return nil, rejectSale("unit %s is not a unit of product %s", unitId.Hex(), productId.Hex())
		}

		rung := sale.Ring(sale.Line{
			ProductId:     line.ProductId,
			UnitId:        line.UnitId,
			Quantity:      line.Quantity,
			PriceType:     line.PriceType,
			StockId:       line.StockId,
			Discount:      line.Discount,
			AllowOversell: line.AllowOversell,
		}, u.catalog)

		drawn := make([]entities.OrderItemStock, 0, len(rung.Parts))
		for i, part := range rung.Parts {
			if part.StockId == "" {
				res, err := entity.productsRepo.UpdateOne(ctx, bson.M{"_id": productId}, bson.M{"$inc": bson.M{"soldFirst": -part.Quantity}})
				if err != nil {
					return nil, err
				}
				if res.MatchedCount == 0 {
					return nil, rejectSale("product %s not found", productId.Hex())
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
			stock, err := entity.drawStock(ctx, part.StockId, take)
			if err != nil {
				return nil, err
			}
			for j := range u.catalog.Stocks {
				if u.catalog.Stocks[j].Id == stock.Id {
					u.catalog.Stocks[j].Quantity = stock.Quantity
				}
			}
			if _, seen := updated[part.StockId]; !seen {
				updatedOrder = append(updatedOrder, part.StockId)
			}
			updated[part.StockId] = *stock
			drawn = append(drawn, entities.OrderItemStock{StockId: part.StockId, Quantity: take})
		}

		items = append(items, entities.OrderItem{
			Id:          primitive.NewObjectID(),
			BranchId:    branchId,
			OrderId:     orderId,
			ProductId:   productId,
			UnitId:      unitId,
			Status:      constant.CONFIRMED,
			Stocks:      drawn,
			Quantity:    line.Quantity,
			Price:       rung.Amount,
			CostPrice:   rung.Cost,
			Discount:    rung.Discount,
			OversoldQty: rung.Oversold,
			CreatedBy:   form.CreatedBy,
			CreatedDate: now,
			UpdatedBy:   form.CreatedBy,
			UpdatedDate: now,
		})
		total += rung.Paid(line.Quantity)
		totalCost += rung.Cost
		discount += rung.Discount * float64(line.Quantity)

		balance, err := entity.getProductStockBalanceWithContext(ctx, productId, unitId, form.BranchId)
		if err != nil {
			return nil, err
		}
		h := request.AddOrderItemProductHistory(line.ProductId, u.unit.Unit,
			request.OrderItem{Quantity: line.Quantity, Price: rung.Amount, CostPrice: rung.Cost}, balance, form.CreatedBy)
		if _, err := entity.productHistoryRepo.InsertOne(ctx, entities.ProductHistory{
			Id:          primitive.NewObjectID(),
			BranchId:    branchId,
			ProductId:   productId,
			Type:        h.Type,
			Description: h.Description,
			Unit:        h.Unit,
			Quantity:    h.Quantity,
			CostPrice:   h.CostPrice,
			Price:       h.Price,
			Balance:     h.Balance,
			CreatedBy:   h.CreatedBy,
			CreatedDate: now,
		}); err != nil {
			return nil, err
		}
	}
	total, totalCost, discount = roundMoney(total), roundMoney(totalCost), roundMoney(discount)

	var tendered float64
	for _, p := range form.Payments {
		if p.Amount < 0 {
			return nil, rejectSale("payment amount must not be negative")
		}
		tendered += p.Amount
	}
	if roundMoney(tendered) < total {
		return nil, rejectSale("payment %.2f is less than the total %.2f", tendered, total)
	}
	change := roundMoney(tendered - total)

	order := entities.Order{
		Id:              orderId,
		BranchId:        branchId,
		Code:            form.Code,
		CustomerCode:    form.CustomerCode,
		CustomerName:    form.CustomerName,
		PatientId:       form.PatientId,
		PharmacistName:  form.PharmacistName,
		LicenseNo:       form.LicenseNo,
		PrescriberName:  form.PrescriberName,
		BuyerName:       form.BuyerName,
		BuyerIdCard:     form.BuyerIdCard,
		Status:          constant.CONFIRMED,
		Total:           total,
		TotalCost:       totalCost,
		Discount:        discount,
		Type:            form.Type,
		SaleId:          form.SaleId,
		SaleFingerprint: saleFingerprint(form),
		CreatedBy:       form.CreatedBy,
		CreatedDate:     now,
		UpdatedBy:       form.CreatedBy,
		UpdatedDate:     now,
	}
	if _, err := entity.orderRepo.InsertOne(ctx, order); err != nil {
		return nil, err
	}
	if _, err := entity.orderItemRepo.InsertMany(ctx, items); err != nil {
		return nil, err
	}
	payments := make([]interface{}, len(form.Payments))
	for i, p := range form.Payments {
		payment := entities.Payment{
			Id:          primitive.NewObjectID(),
			BranchId:    branchId,
			OrderId:     orderId,
			Status:      constant.ACTIVE,
			Amount:      p.Amount,
			Total:       total,
			Change:      change,
			Type:        p.Type,
			CreatedBy:   form.CreatedBy,
			CreatedDate: now,
			UpdatedBy:   form.CreatedBy,
			UpdatedDate: now,
		}
		payments[i] = payment
		order.Payments = append(order.Payments, payment)
	}
	if _, err := entity.paymentRepo.InsertMany(ctx, payments); err != nil {
		return nil, err
	}

	stocks := make([]entities.ProductStock, 0, len(updatedOrder))
	for _, id := range updatedOrder {
		stocks = append(stocks, updated[id])
	}
	return &RecordedSale{Order: &order, Stocks: stocks}, nil
}

// drawStock takes quantity from a Stock that holds at least that much.
func (entity *orderEntity) drawStock(ctx context.Context, stockId string, quantity int) (*entities.ProductStock, error) {
	id, err := primitive.ObjectIDFromHex(stockId)
	if err != nil {
		return nil, err
	}
	var stock entities.ProductStock
	err = entity.productStockRepo.FindOneAndUpdate(ctx,
		bson.M{"_id": id, "quantity": bson.M{"$gte": quantity}},
		bson.M{"$inc": bson.M{"quantity": -quantity}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&stock)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Read in this transaction as holding enough; only a concurrent
		// change can get here, and that aborts the transaction.
		return nil, fmt.Errorf("stock %s changed while recording the sale", stockId)
	}
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func roundMoney(v float64) float64 { return math.Round(v*100) / 100 }
