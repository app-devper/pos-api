package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Order analytics: the dashboard's reads over recorded Orders and Lines.
// They never write, so they live apart from the Order repository, whose
// interface is what a Sale, a cancel and the order screens use.

// IOrderAnalytics is what the dashboard reads about Orders.
type IOrderAnalytics interface {
	GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error)
	GetOrderDailyChart(form request.GetOrderRange) ([]entities.OrderDailyChart, error)
	GetOrderMonthlyChart(branchId string) ([]entities.OrderDailyChart, error)
	GetABCAnalysis(branchId string) ([]entities.ABCProduct, error)
}

type orderAnalyticsEntity struct {
	orderRepo     *mongo.Collection
	orderItemRepo *mongo.Collection
}

func NewOrderAnalyticsEntity(resource *db.Resource) IOrderAnalytics {
	return &orderAnalyticsEntity{
		orderRepo:     resource.PosDb.Collection("orders"),
		orderItemRepo: resource.PosDb.Collection("order_items"),
	}
}

// lineDiscountExpr is a Line's discount: discount is per unit.
func lineDiscountExpr() bson.M {
	return bson.M{"$multiply": bson.A{bson.M{"$ifNull": bson.A{"$discount", 0}}, "$quantity"}}
}

// lineTotalExpr is what a Line charges: price is already the Line amount
// (unit price × quantity), before its per-unit discount.
func lineTotalExpr() bson.M {
	return bson.M{"$subtract": bson.A{"$price", lineDiscountExpr()}}
}

func (entity *orderAnalyticsEntity) GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error) {
	logrus.Info("GetOrderSummary")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchFilter, err := buildActiveOrderAnalyticsMatchFilter(form.StartDate.Time, form.EndDate.Time, form.BranchId)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id":          nil,
			"totalOrders":  bson.M{"$sum": 1},
			"totalRevenue": bson.M{"$sum": "$total"},
			"totalCost":    bson.M{"$sum": "$totalCost"},
		}},
		{"$addFields": bson.M{
			"totalProfit": bson.M{"$subtract": bson.A{"$totalRevenue", "$totalCost"}},
		}},
	}

	var results []entities.OrderSummary
	cursor, err := entity.orderRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &entities.OrderSummary{}, nil
	}
	return &results[0], nil
}

func (entity *orderAnalyticsEntity) GetOrderDailyChart(form request.GetOrderRange) ([]entities.OrderDailyChart, error) {
	logrus.Info("GetOrderDailyChart")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchFilter, err := buildActiveOrderAnalyticsMatchFilter(form.StartDate.Time, form.EndDate.Time, form.BranchId)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$createdDate"},
			},
			"totalOrders":  bson.M{"$sum": 1},
			"totalRevenue": bson.M{"$sum": "$total"},
			"totalCost":    bson.M{"$sum": "$totalCost"},
		}},
		{"$addFields": bson.M{
			"totalProfit": bson.M{"$subtract": bson.A{"$totalRevenue", "$totalCost"}},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	var results []entities.OrderDailyChart
	cursor, err := entity.orderRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.OrderDailyChart{}
	}
	return results, nil
}

func (entity *orderAnalyticsEntity) GetOrderMonthlyChart(branchId string) ([]entities.OrderDailyChart, error) {
	logrus.Info("GetOrderMonthlyChart")
	ctx, cancel := utils.InitContext()
	defer cancel()

	startDate := time.Now().AddDate(-1, 0, 0)
	matchFilter, err := buildMonthlyActiveOrderAnalyticsMatchFilter(startDate, branchId)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m", "date": "$createdDate"},
			},
			"totalOrders":  bson.M{"$sum": 1},
			"totalRevenue": bson.M{"$sum": "$total"},
			"totalCost":    bson.M{"$sum": "$totalCost"},
		}},
		{"$addFields": bson.M{
			"totalProfit": bson.M{"$subtract": bson.A{"$totalRevenue", "$totalCost"}},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	var results []entities.OrderDailyChart
	cursor, err := entity.orderRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.OrderDailyChart{}
	}
	return results, nil
}

func buildMonthlyActiveOrderAnalyticsMatchFilter(startDate time.Time, branchId string) (bson.M, error) {
	matchFilter := bson.M{
		"createdDate": bson.M{"$gte": startDate},
		"status":      bson.M{"$in": constant.ConfirmedOrderStatuses()},
	}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjId
	}
	return matchFilter, nil
}

func buildActiveOrderAnalyticsMatchFilter(startDate time.Time, endDate time.Time, branchId string) (bson.M, error) {
	matchFilter := bson.M{
		"createdDate": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
		"status": bson.M{"$in": constant.ConfirmedOrderStatuses()},
	}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjId
	}
	return matchFilter, nil
}

func (entity *orderAnalyticsEntity) GetABCAnalysis(branchId string) ([]entities.ABCProduct, error) {
	logrus.Info("GetABCAnalysis")
	ctx, cancel := utils.InitContext()
	defer cancel()

	startDate := time.Now().AddDate(0, -3, 0)
	pipeline, err := buildABCAnalysisPipeline(startDate, branchId)
	if err != nil {
		return nil, err
	}

	var abcResults []entities.ABCProduct
	cursor, err := entity.orderItemRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &abcResults); err != nil {
		return nil, err
	}
	if abcResults == nil {
		abcResults = []entities.ABCProduct{}
	}

	var totalRev float64
	for _, r := range abcResults {
		totalRev += r.TotalRevenue
	}
	if totalRev > 0 {
		var cumRev float64
		for i := range abcResults {
			cumRev += abcResults[i].TotalRevenue
			pct := cumRev / totalRev
			if pct <= 0.80 {
				abcResults[i].Class = "A"
			} else if pct <= 0.95 {
				abcResults[i].Class = "B"
			} else {
				abcResults[i].Class = "C"
			}
		}
	}

	return abcResults, nil
}

func buildABCAnalysisPipeline(startDate time.Time, branchId string) ([]bson.M, error) {
	matchFilter := bson.M{
		"createdDate": bson.M{"$gte": startDate},
		"$or":         ledger.StandingLines(),
	}
	orderMatch := bson.A{
		bson.M{"$eq": bson.A{"$_id", "$$oid"}},
		bson.M{"$in": bson.A{"$status", bson.A{constant.ACTIVE, constant.CONFIRMED}}},
	}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjId
		orderMatch = append(orderMatch, bson.M{"$eq": bson.A{"$branchId", branchObjId}})
	}

	return []bson.M{
		{"$match": matchFilter},
		{"$lookup": bson.M{
			"from": "orders",
			"let":  bson.M{"oid": "$orderId"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": orderMatch}}},
				bson.M{"$project": bson.M{"_id": 1}},
			},
			"as": "order",
		}},
		{"$match": bson.M{"order.0": bson.M{"$exists": true}}},
		{"$group": bson.M{
			"_id":          "$productId",
			"totalRevenue": bson.M{"$sum": lineTotalExpr()},
			"totalQty":     bson.M{"$sum": "$quantity"},
		}},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": bson.M{"path": "$product", "preserveNullAndEmptyArrays": true}},
		{"$addFields": bson.M{
			"productName": bson.M{"$ifNull": bson.A{"$product.name", ""}},
		}},
		{"$project": bson.M{"product": 0}},
		{"$sort": bson.M{"totalRevenue": -1}},
	}, nil
}
