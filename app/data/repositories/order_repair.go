package repositories

import (
	"context"
	"math"

	"pos/app/domain/constant"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrderTotalsRepair is an Order whose stored money differs from its Lines.
type OrderTotalsRepair struct {
	OrderId                             primitive.ObjectID
	Code                                string
	Total, TotalCost, Discount          float64 // stored
	NewTotal, NewTotalCost, NewDiscount float64 // from its confirmed Lines
}

// RepairCancelledLineTotals finds Orders still standing that have a
// cancelled Line, and whose total, totalCost or discount differ from their
// confirmed Lines. Before the fix a cancelled Line recomputed them as if a
// Line's price were per unit. With apply it rewrites them; otherwise it only
// reports. Repeating it changes nothing.
func RepairCancelledLineTotals(ctx context.Context, pos *mongo.Database, apply bool) ([]OrderTotalsRepair, error) {
	e := newOrderEntity(&db.Resource{PosDb: pos})
	orderIds, err := e.orderItemRepo.Distinct(ctx, "orderId", bson.M{"status": constant.CANCELLED})
	if err != nil {
		return nil, err
	}
	var repairs []OrderTotalsRepair
	for _, raw := range orderIds {
		id, ok := raw.(primitive.ObjectID)
		if !ok {
			continue
		}
		var order struct {
			Code      string  `bson:"code"`
			Status    string  `bson:"status"`
			Total     float64 `bson:"total"`
			TotalCost float64 `bson:"totalCost"`
			Discount  float64 `bson:"discount"`
		}
		if err := e.orderRepo.FindOne(ctx, bson.M{"_id": id}).Decode(&order); err != nil {
			return nil, err
		}
		if order.Status == constant.CANCELLED {
			continue
		}
		totals, err := e.getOrderTotalsWithContext(ctx, id.Hex())
		if err != nil {
			return nil, err
		}
		totals = orderTotals{total: roundMoney(totals.total), totalCost: roundMoney(totals.totalCost), discount: roundMoney(totals.discount)}
		if same(order.Total, totals.total) && same(order.TotalCost, totals.totalCost) && same(order.Discount, totals.discount) {
			continue
		}
		repairs = append(repairs, OrderTotalsRepair{
			OrderId: id, Code: order.Code,
			Total: order.Total, TotalCost: order.TotalCost, Discount: order.Discount,
			NewTotal: totals.total, NewTotalCost: totals.totalCost, NewDiscount: totals.discount,
		})
		if apply {
			if _, err := e.updateTotalOrderByIdWithContext(ctx, id.Hex()); err != nil {
				return repairs, err
			}
		}
	}
	return repairs, nil
}

func same(a, b float64) bool { return math.Abs(a-b) < 0.005 }
