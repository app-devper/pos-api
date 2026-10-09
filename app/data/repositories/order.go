package repositories

import (
	"context"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/sale"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type orderEntity struct {
	client           *mongo.Client
	ledger           *ledger.Ledger
	orderRepo        *mongo.Collection
	orderItemRepo    *mongo.Collection
	paymentRepo      *mongo.Collection
	productStockRepo *mongo.Collection
}

type IOrder interface {
	RecordSale(form request.Sale) (*RecordedSale, error)
	GetOrderRange(form request.GetOrderRange) ([]entities.Order, error)
	GetOrdersByCustomerCode(customerCode string, branchId string) ([]entities.Order, error)
	GetOrderById(id string, branchId string) (*entities.Order, error)
	GetOrderDetailById(id string) (*entities.OrderDetail, error)
	UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error)
	CancelOrderById(id string, userId string, branchId string, reason string) (*entities.OrderDetail, error)

	GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error)
	GetOrderItemById(id string, branchId string) (*entities.OrderItem, error)
	CancelOrderItemById(id string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error)
	GetOrderItemDetailById(id string) (*entities.OrderItemProductDetail, error)
	CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error)
	GetOrderItemByProductId(productId string, branchId string) ([]entities.OrderItem, error)
	GetOrderItemOrderDetailsByProductId(productId string, branchId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error)
}

func NewOrderEntity(resource *db.Resource) IOrder {
	entity := newOrderEntity(resource)
	ensureOrderIndexes(entity.orderRepo, entity.orderItemRepo, entity.paymentRepo)
	return entity
}

func newOrderEntity(resource *db.Resource) *orderEntity {
	orderRepo := resource.PosDb.Collection("orders")
	orderItemRepo := resource.PosDb.Collection("order_items")
	paymentRepo := resource.PosDb.Collection("payments")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	entity := &orderEntity{
		client: resource.Client, ledger: newLedger(resource), orderRepo: orderRepo, orderItemRepo: orderItemRepo, paymentRepo: paymentRepo, productStockRepo: productStockRepo,
	}
	return entity
}

func ensureOrderIndexes(orderRepo *mongo.Collection, orderItemRepo *mongo.Collection, paymentRepo *mongo.Collection) {
	createCollectionIndex(orderRepo, "orders createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(orderRepo, "orders saleId", mongo.IndexModel{
		Keys: bson.D{{Key: "saleId", Value: 1}},
		Options: options.Index().SetUnique(true).
			SetPartialFilterExpression(bson.M{"saleId": bson.M{"$type": "string"}}),
	})
	createCollectionIndex(orderRepo, "orders customerCode", mongo.IndexModel{
		Keys: bson.D{{Key: "customerCode", Value: 1}},
	})
	createCollectionIndex(orderItemRepo, "order_items orderId", mongo.IndexModel{
		Keys: bson.D{{Key: "orderId", Value: 1}},
	})
	createCollectionIndex(orderItemRepo, "order_items productId", mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	createCollectionIndex(orderItemRepo, "order_items createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(paymentRepo, "payments orderId", mongo.IndexModel{
		Keys: bson.D{{Key: "orderId", Value: 1}},
	})
	createCollectionIndex(orderRepo, "orders branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(orderItemRepo, "order_items branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(paymentRepo, "payments branchId+orderId", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "orderId", Value: 1}},
	})
	// The Stock ledger looks up the Lines still owed a Unit, oldest first, on
	// every Stock rise (ADR-0002); only owed Lines are indexed.
	createCollectionIndex(orderItemRepo, "order_items owed by branch+product+unit", mongo.IndexModel{
		Keys:    bson.D{{Key: "branchId", Value: 1}, {Key: "productId", Value: 1}, {Key: "unitId", Value: 1}, {Key: "_id", Value: 1}},
		Options: options.Index().SetPartialFilterExpression(bson.M{"oversoldQty": bson.M{"$gt": 0}}),
	})
}

func (entity *orderEntity) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	logrus.Info("GetOrderRange")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{"createdDate": bson.M{
		"$gte": form.StartDate.Time,
		"$lt":  form.EndDate.Time,
	}}
	if form.BranchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(form.BranchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjId
	}
	var items []entities.Order
	cursor, err := entity.orderRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	items = []entities.Order{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if err = entity.populateOrderPayments(items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) GetOrdersByCustomerCode(customerCode string, branchId string) ([]entities.Order, error) {
	logrus.Info("GetOrdersByCustomerCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	var items []entities.Order
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})
	filter := bson.M{"customerCode": customerCode}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	cursor, err := entity.orderRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items = []entities.Order{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if err = entity.populateOrderPayments(items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) populateOrderPayments(items []entities.Order) error {
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.populateOrderPaymentsWithContext(ctx, items)
}

func (entity *orderEntity) populateOrderPaymentsWithContext(ctx context.Context, items []entities.Order) error {
	if len(items) == 0 {
		return nil
	}

	orderIDs := make([]primitive.ObjectID, 0, len(items))
	for i := range items {
		orderIDs = append(orderIDs, items[i].Id)
	}

	paymentMap, err := entity.getPaymentsByOrderIDsWithContext(ctx, orderIDs)
	if err != nil {
		return err
	}

	for i := range items {
		items[i].Payments = paymentMap[items[i].Id.Hex()]
	}
	return nil
}

func (entity *orderEntity) getPaymentsByOrderIDsWithContext(ctx context.Context, orderIDs []primitive.ObjectID) (map[string][]entities.Payment, error) {
	paymentMap := make(map[string][]entities.Payment, len(orderIDs))
	if len(orderIDs) == 0 {
		return paymentMap, nil
	}

	cursor, err := entity.paymentRepo.Find(
		ctx,
		bson.M{"orderId": bson.M{"$in": orderIDs}},
		options.Find().SetSort(bson.D{{Key: "createdDate", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}

	payments := []entities.Payment{}
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}

	for _, payment := range payments {
		key := payment.OrderId.Hex()
		paymentMap[key] = append(paymentMap[key], payment)
	}
	return paymentMap, nil
}

func (entity *orderEntity) GetOrderById(id string, branchId string) (*entities.Order, error) {
	logrus.Info("GetOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter, err := inBranch(id, branchId)
	if err != nil {
		return nil, err
	}
	var data entities.Order
	err = entity.orderRepo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error) {
	logrus.Info("UpdateCustomerCodeOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Order
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"customerCode": customerCode,
		"updatedDate":  time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) GetOrderDetailById(id string) (*entities.OrderDetail, error) {
	logrus.Info("GetOrderDetailById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.OrderDetail
	err = entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	payments, err := entity.GetPaymentsByOrderId(id)
	if err != nil {
		return nil, err
	}
	data.Payments = payments
	if len(payments) > 0 {
		data.Payment = payments[0]
	}

	items, err := entity.getOrderItemDetailByOrderId(id)
	if err != nil {
		return nil, err
	}
	data.Items = items

	return &data, nil
}

// CancelOrderById is recorded by the Stock ledger (ADR-0001); the response is
// CancelOrderById is recorded by the Stock ledger (ADR-0001). The Order is
// read first, so a cancel that committed is never reported as failed.
func (entity *orderEntity) CancelOrderById(id string, userId string, branchId string, reason string) (*entities.OrderDetail, error) {
	logrus.Info("CancelOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	order, err := entity.getOrderDetailByIdWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := entity.ledger.CancelOrder(context.Background(), id, branchId, userId, reason); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"orderId": id, "userId": userId, "branchId": branchId}).Error("cancel order failed")
		return nil, err
	}
	now := time.Now()
	order.Status, order.CancelReason = constant.CANCELLED, reason
	for i := range order.Items {
		order.Items[i].Status, order.Items[i].CancelReason, order.Items[i].UpdatedBy, order.Items[i].UpdatedDate = constant.CANCELLED, reason, userId, now
	}
	for i := range order.Payments {
		order.Payments[i].Status, order.Payments[i].CancelReason, order.Payments[i].UpdatedBy, order.Payments[i].UpdatedDate = constant.CANCELLED, reason, userId, now
	}
	if len(order.Payments) > 0 {
		order.Payment = order.Payments[0]
	}
	return order, nil
}

func (entity *orderEntity) GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemRange")
	ctx, cancel := utils.InitContext()
	defer cancel()
	matchFilter := bson.M{
		"createdDate": bson.M{
			"$gte": form.StartDate.Time,
			"$lt":  form.EndDate.Time,
		},
		"$or": ledger.StandingLines(),
	}
	if form.BranchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(form.BranchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjId
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{"$match": matchFilter},
		{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "productId",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		{"$unwind": "$product"},
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemById(id string, branchId string) (*entities.OrderItem, error) {
	logrus.Info("GetOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter, err := inBranch(id, branchId)
	if err != nil {
		return nil, err
	}
	var data entities.OrderItem
	err = entity.orderItemRepo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// CancelOrderItemById is recorded by the Stock ledger (ADR-0001); the response
// CancelOrderItemById is recorded by the Stock ledger (ADR-0001). The Line is
// read first, so a cancel that committed is never reported as failed.
func (entity *orderEntity) CancelOrderItemById(id string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("CancelOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	item, err := entity.getOrderItemDetailByIdWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := entity.ledger.CancelLine(context.Background(), id, branchId, userId, reason); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"orderItemId": id, "userId": userId, "branchId": branchId}).Error("cancel order item failed")
		return nil, err
	}
	item.Status, item.CancelReason, item.UpdatedBy, item.UpdatedDate = constant.CANCELLED, reason, userId, time.Now()
	return item, nil
}

func (entity *orderEntity) GetOrderItemDetailById(id string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemDetailById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"_id": objId,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "productId",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		{"$unwind": "$product"},
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &items[0], nil
}

func (entity *orderEntity) getOrderItemDetailByOrderId(orderId string) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"orderId": objId,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "productId",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		{"$unwind": "$product"},
	})

	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) getOrderItemDetailByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("getOrderItemDetailByOrderProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	productObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"orderId":   objId,
				"productId": productObjId,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "productId",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		{"$unwind": "$product"},
	})

	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &items[0], nil
}

func (entity *orderEntity) GetOrderItemByProductId(productId string, branchId string) ([]entities.OrderItem, error) {
	logrus.Info("GetOrderItemByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"productId": objId, "$or": ledger.StandingLines()}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	cursor, err := entity.orderItemRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItem{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemOrderDetailsByProductId(productId string, branchId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error) {
	logrus.Info("GetOrderItemOrderDetailsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	productObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	matchFilter := bson.M{
		"productId": productObjId,
		"createdDate": bson.M{
			"$gte": form.StartDate,
			"$lt":  form.EndDate,
		},
		"$or": ledger.StandingLines(),
	}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjID
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$lookup": bson.M{
				"from":         "orders",
				"localField":   "orderId",
				"foreignField": "_id",
				"as":           "order",
			},
		},
		{"$unwind": "$order"},
	})

	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemOrderDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) RemoveOrderItemByOrderId(orderId string) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("RemoveOrderItemByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	items, err := entity.getOrderItemDetailByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	_, err = entity.orderItemRepo.DeleteMany(ctx, bson.M{"orderId": objId})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// CancelOrderItemByOrderProductId cancels an Order's Line of a Product through
// the Stock ledger.
func (entity *orderEntity) CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("CancelOrderItemByOrderProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	item, err := entity.getOrderItemDetailByOrderProductIdWithContext(ctx, orderId, productId)
	if err != nil {
		return nil, err
	}
	return entity.CancelOrderItemById(item.Id.Hex(), userId, branchId, reason)
}

func (entity *orderEntity) GetPaymentsByOrderId(orderId string) ([]entities.Payment, error) {
	logrus.Info("GetPaymentsByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.paymentRepo.Find(ctx, bson.M{"orderId": objId}, options.Find().SetSort(bson.D{{Key: "createdDate", Value: 1}}))
	if err != nil {
		return nil, err
	}
	payments := []entities.Payment{}
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}
	return payments, nil
}

func (entity *orderEntity) RemovePaymentsByOrderId(orderId string) ([]entities.Payment, error) {
	logrus.Info("RemovePaymentsByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	payments, err := entity.GetPaymentsByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	if _, err = entity.paymentRepo.DeleteMany(ctx, bson.M{"orderId": objId}); err != nil {
		return nil, err
	}
	return payments, nil
}

func (entity *orderEntity) getOrderItemDetailByIdWithContext(ctx context.Context, id string) (*entities.OrderItemProductDetail, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"_id": objId}},
		{"$lookup": bson.M{"from": "products", "localField": "productId", "foreignField": "_id", "as": "product"}},
		{"$unwind": "$product"},
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &items[0], nil
}

func (entity *orderEntity) getOrderDetailByIdWithContext(ctx context.Context, id string) (*entities.OrderDetail, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.OrderDetail
	if err = entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data); err != nil {
		return nil, err
	}

	payments := []entities.Payment{}
	cursor, err := entity.paymentRepo.Find(ctx, bson.M{"orderId": objId}, options.Find().SetSort(bson.D{{Key: "createdDate", Value: 1}}))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}
	data.Payments = payments
	if len(payments) > 0 {
		data.Payment = payments[0]
	}

	items, err := entity.getOrderItemDetailsByOrderIdWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	data.Items = items
	return &data, nil
}

func (entity *orderEntity) getOrderItemDetailByOrderProductIdWithContext(ctx context.Context, orderId string, productId string) (*entities.OrderItemProductDetail, error) {
	orderObjId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	productObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"orderId": orderObjId, "productId": productObjId, "$or": ledger.StandingLines()}},
		{"$sort": bson.M{"_id": 1}},
		{"$lookup": bson.M{"from": "products", "localField": "productId", "foreignField": "_id", "as": "product"}},
		{"$unwind": "$product"},
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &items[0], nil
}

func (entity *orderEntity) getOrderItemDetailsByOrderIdWithContext(ctx context.Context, orderId string) ([]entities.OrderItemProductDetail, error) {
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"orderId": objId}},
		{"$lookup": bson.M{"from": "products", "localField": "productId", "foreignField": "_id", "as": "product"}},
		{"$unwind": "$product"},
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItemProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) updateTotalOrderByIdWithContext(ctx context.Context, id string) (*entities.Order, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	totals, err := entity.getOrderTotalsWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}
	var data entities.Order
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"total":       totals.total,
		"totalCost":   totals.totalCost,
		"discount":    totals.discount,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// orderTotals are an Order's money summed over its standing Lines.
type orderTotals struct {
	total, totalCost, discount float64
}

// getOrderTotalsWithContext is the Order's money from its standing Lines,
// by the same rule the Stock ledger recomputes with (sale.OrderMoney).
func (entity *orderEntity) getOrderTotalsWithContext(ctx context.Context, orderId string) (orderTotals, error) {
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return orderTotals{}, err
	}
	cursor, err := entity.orderItemRepo.Find(ctx, bson.M{"orderId": objId, "$or": ledger.StandingLines()})
	if err != nil {
		return orderTotals{}, err
	}
	var lines []entities.OrderItem
	if err = cursor.All(ctx, &lines); err != nil {
		return orderTotals{}, err
	}
	total, cost, discount := sale.OrderMoney(lines)
	return orderTotals{total: total, totalCost: cost, discount: discount}, nil
}
