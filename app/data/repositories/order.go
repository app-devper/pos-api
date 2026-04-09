package repositories

import (
	"context"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type orderEntity struct {
	client             *mongo.Client
	orderRepo          *mongo.Collection
	orderItemRepo      *mongo.Collection
	paymentRepo        *mongo.Collection
	productsRepo       *mongo.Collection
	productStockRepo   *mongo.Collection
	productUnitsRepo   *mongo.Collection
	productHistoryRepo *mongo.Collection
}

type IOrder interface {
	CreateOrder(form request.Order) (*entities.Order, error)
	GetOrderRange(form request.GetOrderRange) ([]entities.Order, error)
	GetOrdersByCustomerCode(customerCode string, branchId string) ([]entities.Order, error)
	UpdateTotal() ([]entities.Order, error)
	GetOrderById(id string) (*entities.Order, error)
	GetOrderDetailById(id string) (*entities.OrderDetail, error)
	UpdateTotalCostOrderById(id string, totalCost float64) (*entities.Order, error)
	UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error)
	RemoveOrderById(id string) (*entities.OrderDetail, error)
	CancelOrderById(id string, userId string, branchId string) (*entities.OrderDetail, error)
	UpdateTotalOrderById(id string) (*entities.Order, error)
	GetTotalOrderById(id string) float64
	GetTotalCostOrderById(id string) float64

	GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error)
	GetOrderItemById(id string) (*entities.OrderItem, error)
	UpdateOrderItemById(id string, form request.OrderItem) (*entities.OrderItem, error)
	RemoveOrderItemById(id string) (*entities.OrderItemProductDetail, error)
	CancelOrderItemById(id string, userId string, branchId string) (*entities.OrderItemProductDetail, error)
	GetOrderItemDetailById(id string) (*entities.OrderItemProductDetail, error)
	GetOrderItemDetailByOrderId(orderId string) ([]entities.OrderItemProductDetail, error)
	GetOrderItemDetailByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error)
	RemoveOrderItemByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error)
	CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error)
	GetOrderItemByProductId(productId string) ([]entities.OrderItem, error)
	GetOrderItemOrderDetailsByProductId(productId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error)

	GetPaymentByOrderId(orderId string) (*entities.Payment, error)
	RemovePaymentByOrderId(orderId string) (*entities.Payment, error)

	GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error)
	GetOrderDailyChart(form request.GetOrderRange) ([]entities.OrderDailyChart, error)
	GetOrderMonthlyChart(branchId string) ([]entities.OrderDailyChart, error)
	GetABCAnalysis(branchId string) ([]entities.ABCProduct, error)
}

func NewOrderEntity(resource *db.Resource) IOrder {
	orderRepo := resource.PosDb.Collection("orders")
	orderItemRepo := resource.PosDb.Collection("order_items")
	paymentRepo := resource.PosDb.Collection("payments")
	productsRepo := resource.PosDb.Collection("products")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	productUnitsRepo := resource.PosDb.Collection("product_units")
	productHistoryRepo := resource.PosDb.Collection("product_histories")
	entity := &orderEntity{
		client: resource.Client, orderRepo: orderRepo, orderItemRepo: orderItemRepo, paymentRepo: paymentRepo,
		productsRepo: productsRepo, productStockRepo: productStockRepo, productUnitsRepo: productUnitsRepo, productHistoryRepo: productHistoryRepo,
	}
	ensureOrderIndexes(orderRepo, orderItemRepo, paymentRepo)
	return entity
}

func ensureOrderIndexes(orderRepo *mongo.Collection, orderItemRepo *mongo.Collection, paymentRepo *mongo.Collection) {
	createCollectionIndex(orderRepo, "orders createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
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
}

func (entity *orderEntity) CreateOrder(form request.Order) (*entities.Order, error) {
	logrus.Info("CreateOrder")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createOrderWithContext(ctx, form)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var created *entities.Order
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		order, txErr := entity.createOrderWithContext(sessCtx, form)
		if txErr != nil {
			return nil, txErr
		}
		created = order
		return order, nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (entity *orderEntity) createOrderWithContext(ctx context.Context, form request.Order) (*entities.Order, error) {
	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
	var orderId = primitive.NewObjectID()
	data := entities.Order{
		Id:             orderId,
		BranchId:       branchId,
		Code:           form.Code,
		CustomerCode:   form.CustomerCode,
		CustomerName:   form.CustomerName,
		PatientId:      form.PatientId,
		PharmacistName: form.PharmacistName,
		LicenseNo:      form.LicenseNo,
		PrescriberName: form.PrescriberName,
		BuyerName:      form.BuyerName,
		BuyerIdCard:    form.BuyerIdCard,
		Status:         constant.ACTIVE,
		Total:          form.Total,
		TotalCost:      form.TotalCost,
		Discount:       form.Discount,
		Type:           form.Type,
		CreatedBy:      form.CreatedBy,
		CreatedDate:    time.Now(),
		UpdatedBy:      form.CreatedBy,
		UpdatedDate:    time.Now(),
	}
	_, err = entity.orderRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}

	count := len(form.Items)
	orderItem := make([]interface{}, count)
	for i := 0; i < count; i++ {
		formItem := form.Items[i]
		productId, err := primitive.ObjectIDFromHex(formItem.ProductId)
		if err != nil {
			return nil, err
		}
		unitId, err := primitive.ObjectIDFromHex(formItem.UnitId)
		if err != nil {
			return nil, err
		}
		countStock := len(formItem.Stocks)
		stocks := make([]entities.OrderItemStock, countStock)
		for j := 0; j < countStock; j++ {
			formStock := formItem.Stocks[j]
			stock := entities.OrderItemStock{
				Quantity: formStock.Quantity,
				StockId:  formStock.StockId,
			}
			stocks[j] = stock
		}
		item := entities.OrderItem{
			Id:          primitive.NewObjectID(),
			BranchId:    branchId,
			OrderId:     orderId,
			ProductId:   productId,
			UnitId:      unitId,
			Stocks:      stocks,
			Quantity:    formItem.Quantity,
			Price:       formItem.Price,
			CostPrice:   formItem.CostPrice,
			Discount:    formItem.Discount,
			CreatedBy:   form.CreatedBy,
			CreatedDate: time.Now(),
			UpdatedBy:   form.CreatedBy,
			UpdatedDate: time.Now(),
		}
		orderItem[i] = item
	}
	_, err = entity.orderItemRepo.InsertMany(ctx, orderItem)
	if err != nil {
		return nil, err
	}

	if len(form.Payments) > 0 {
		payments := make([]interface{}, len(form.Payments))
		for i, p := range form.Payments {
			payments[i] = entities.Payment{
				Id:          primitive.NewObjectID(),
				BranchId:    branchId,
				OrderId:     orderId,
				Status:      constant.ACTIVE,
				Amount:      p.Amount,
				Total:       form.Total,
				Change:      form.Change,
				Type:        p.Type,
				CreatedBy:   form.CreatedBy,
				CreatedDate: time.Now(),
				UpdatedBy:   form.CreatedBy,
				UpdatedDate: time.Now(),
			}
		}
		_, err = entity.paymentRepo.InsertMany(ctx, payments)
	} else {
		payment := entities.Payment{
			Id:          primitive.NewObjectID(),
			BranchId:    branchId,
			OrderId:     orderId,
			Status:      constant.ACTIVE,
			Amount:      form.Amount,
			Total:       form.Total,
			Change:      form.Change,
			Type:        form.Type,
			CreatedBy:   form.CreatedBy,
			CreatedDate: time.Now(),
			UpdatedBy:   form.CreatedBy,
			UpdatedDate: time.Now(),
		}
		_, err = entity.paymentRepo.InsertOne(ctx, payment)
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	logrus.Info("GetOrderRange")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{"createdDate": bson.M{
		"$gte": form.StartDate,
		"$lt":  form.EndDate,
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

func (entity *orderEntity) UpdateTotal() ([]entities.Order, error) {
	logrus.Info("UpdateTotal")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var items []entities.Order
	cursor, err := entity.orderRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var data entities.Order
		err = cursor.Decode(&data)
		if err != nil {
			logrus.WithError(err).WithField("current", cursor.Current.String()).Error("failed to decode order while updating totals")
		} else {
			if data.Total == 0 {
				data.Total = entity.GetTotalOrderById(data.Id.Hex())
				data.TotalCost = entity.GetTotalCostOrderById(data.Id.Hex())
				isReturnNewDoc := options.After
				opts := &options.FindOneAndUpdateOptions{
					ReturnDocument: &isReturnNewDoc,
				}
				err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
				if err != nil {
					return nil, err
				}
			}
			items = append(items, data)
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []entities.Order{}
	}
	return items, nil
}

func (entity *orderEntity) GetOrderById(id string) (*entities.Order, error) {
	logrus.Info("GetOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.Order
	err = entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) UpdateTotalCostOrderById(id string, totalCost float64) (*entities.Order, error) {
	logrus.Info("UpdateTotalCostOrderById")
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
		"totalCost":   totalCost,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
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

	items, err := entity.GetOrderItemDetailByOrderId(id)
	if err != nil {
		return nil, err
	}
	data.Items = items

	return &data, nil
}

func (entity *orderEntity) RemoveOrderById(id string) (*entities.OrderDetail, error) {
	logrus.Info("RemoveOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.OrderDetail
	err = entity.orderRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	payments, pErr := entity.RemovePaymentsByOrderId(id)
	if pErr == nil {
		data.Payments = payments
		if len(payments) > 0 {
			data.Payment = payments[0]
		}
	}

	items, _ := entity.RemoveOrderItemByOrderId(id)
	data.Items = items

	return &data, nil
}

func (entity *orderEntity) CancelOrderById(id string, userId string, branchId string) (*entities.OrderDetail, error) {
	logrus.Info("CancelOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		result, err := entity.cancelOrderByIdWithContext(ctx, id, userId, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"orderId":  id,
				"userId":   userId,
				"branchId": branchId,
			}).Error("cancel order failed")
		}
		return result, err
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.OrderDetail
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.cancelOrderByIdWithContext(sessCtx, id, userId, branchId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"orderId":  id,
			"userId":   userId,
			"branchId": branchId,
		}).Error("cancel order transaction failed")
		return nil, err
	}
	return result, nil
}

func (entity *orderEntity) UpdateTotalOrderById(id string) (*entities.Order, error) {
	logrus.Info("UpdateTotalOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	total, totalCost := entity.getOrderTotals(id)

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Order
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"total":       total,
		"totalCost":   totalCost,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (entity *orderEntity) getOrderTotals(orderId string) (total float64, totalCost float64) {
	logrus.Info("getOrderTotals")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return 0, 0
	}
	pipeline := []bson.M{
		{"$match": bson.M{"orderId": objId}},
		{"$group": bson.M{
			"_id":       nil,
			"total":     bson.M{"$sum": "$price"},
			"totalCost": bson.M{"$sum": "$costPrice"},
		}},
	}
	var result []bson.M
	cursor, err := entity.orderItemRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0
	}
	if err = cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0, 0
	}
	if v, ok := result[0]["total"].(float64); ok {
		total = v
	}
	if v, ok := result[0]["totalCost"].(float64); ok {
		totalCost = v
	}
	return total, totalCost
}

func (entity *orderEntity) GetTotalOrderById(orderId string) float64 {
	total, _ := entity.getOrderTotals(orderId)
	return total
}

func (entity *orderEntity) GetTotalCostOrderById(orderId string) float64 {
	_, totalCost := entity.getOrderTotals(orderId)
	return totalCost
}

func (entity *orderEntity) GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemRange")
	ctx, cancel := utils.InitContext()
	defer cancel()
	matchFilter := bson.M{
		"createdDate": bson.M{
			"$gte": form.StartDate,
			"$lt":  form.EndDate,
		},
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

func (entity *orderEntity) GetOrderItemById(id string) (*entities.OrderItem, error) {
	logrus.Info("GetOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.OrderItem
	err = entity.orderItemRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) UpdateOrderItemById(id string, form request.OrderItem) (*entities.OrderItem, error) {
	logrus.Info("UpdateOrderItemById")
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
	var data entities.OrderItem
	err = entity.orderItemRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"discount":    form.Discount,
		"price":       form.Price,
		"costPrice":   form.CostPrice,
		"quantity":    form.Quantity,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) RemoveOrderItemById(id string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("RemoveOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	item, err := entity.GetOrderItemDetailById(id)
	if err != nil {
		return nil, err
	}
	_, err = entity.orderItemRepo.DeleteOne(ctx, bson.M{"_id": objId})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (entity *orderEntity) CancelOrderItemById(id string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("CancelOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		result, err := entity.cancelOrderItemByIdWithContext(ctx, id, userId, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"orderItemId": id,
				"userId":      userId,
				"branchId":    branchId,
			}).Error("cancel order item failed")
		}
		return result, err
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.OrderItemProductDetail
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.cancelOrderItemByIdWithContext(sessCtx, id, userId, branchId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"orderItemId": id,
			"userId":      userId,
			"branchId":    branchId,
		}).Error("cancel order item transaction failed")
		return nil, err
	}
	return result, nil
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

func (entity *orderEntity) GetOrderItemDetailByOrderId(orderId string) ([]entities.OrderItemProductDetail, error) {
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

func (entity *orderEntity) GetOrderItemDetailByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemDetailByOrderProductId")
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

func (entity *orderEntity) GetOrderItemByProductId(productId string) ([]entities.OrderItem, error) {
	logrus.Info("GetOrderItemByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Find(ctx, bson.M{
		"productId": objId,
	})
	if err != nil {
		return nil, err
	}
	items := []entities.OrderItem{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemOrderDetailsByProductId(productId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error) {
	logrus.Info("GetOrderItemOrderDetailsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	productObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"productId": productObjId,
				"createdDate": bson.M{
					"$gt": form.StartDate,
					"$lt": form.EndDate,
				},
			},
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
	items, err := entity.GetOrderItemDetailByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	_, err = entity.orderItemRepo.DeleteMany(ctx, bson.M{"orderId": objId})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *orderEntity) RemoveOrderItemByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("RemoveOrderItemByOrderProductId")
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
	item, err := entity.GetOrderItemDetailByOrderProductId(orderId, productId)
	if err != nil {
		return nil, err
	}
	_, err = entity.orderItemRepo.DeleteOne(ctx, bson.M{"orderId": objId, "productId": productObjId})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (entity *orderEntity) CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("CancelOrderItemByOrderProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		result, err := entity.cancelOrderItemByOrderProductIdWithContext(ctx, orderId, productId, userId, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"orderId":   orderId,
				"productId": productId,
				"userId":    userId,
				"branchId":  branchId,
			}).Error("cancel order item by product failed")
		}
		return result, err
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.OrderItemProductDetail
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.cancelOrderItemByOrderProductIdWithContext(sessCtx, orderId, productId, userId, branchId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"orderId":   orderId,
			"productId": productId,
			"userId":    userId,
			"branchId":  branchId,
		}).Error("cancel order item by product transaction failed")
		return nil, err
	}
	return result, nil
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

func (entity *orderEntity) GetPaymentByOrderId(orderId string) (*entities.Payment, error) {
	payments, err := entity.GetPaymentsByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &payments[0], nil
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

func (entity *orderEntity) RemovePaymentByOrderId(orderId string) (*entities.Payment, error) {
	payments, err := entity.RemovePaymentsByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &payments[0], nil
}

func (entity *orderEntity) cancelOrderItemByIdWithContext(ctx context.Context, id string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	item, err := entity.getOrderItemDetailByIdWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	if _, err = entity.orderItemRepo.DeleteOne(ctx, bson.M{"_id": objId}); err != nil {
		return nil, err
	}
	if err = entity.restoreOrderItemStockAndHistory(ctx, item, userId, branchId); err != nil {
		return nil, err
	}
	if _, err = entity.updateTotalOrderByIdWithContext(ctx, item.OrderId.Hex()); err != nil {
		return nil, err
	}
	return item, nil
}

func (entity *orderEntity) cancelOrderItemByOrderProductIdWithContext(ctx context.Context, orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	item, err := entity.getOrderItemDetailByOrderProductIdWithContext(ctx, orderId, productId)
	if err != nil {
		return nil, err
	}
	orderObjId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	productObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	if _, err = entity.orderItemRepo.DeleteOne(ctx, bson.M{"orderId": orderObjId, "productId": productObjId}); err != nil {
		return nil, err
	}
	if err = entity.restoreOrderItemStockAndHistory(ctx, item, userId, branchId); err != nil {
		return nil, err
	}
	if _, err = entity.updateTotalOrderByIdWithContext(ctx, orderId); err != nil {
		return nil, err
	}
	return item, nil
}

func (entity *orderEntity) cancelOrderByIdWithContext(ctx context.Context, id string, userId string, branchId string) (*entities.OrderDetail, error) {
	order, err := entity.getOrderDetailByIdWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	for i := range order.Items {
		if err = entity.restoreOrderItemStockAndHistory(ctx, &order.Items[i], userId, branchId); err != nil {
			return nil, err
		}
	}

	if _, err = entity.paymentRepo.DeleteMany(ctx, bson.M{"orderId": objId}); err != nil {
		return nil, err
	}
	if _, err = entity.orderItemRepo.DeleteMany(ctx, bson.M{"orderId": objId}); err != nil {
		return nil, err
	}
	if _, err = entity.orderRepo.DeleteOne(ctx, bson.M{"_id": objId}); err != nil {
		return nil, err
	}
	return order, nil
}

func (entity *orderEntity) restoreOrderItemStockAndHistory(ctx context.Context, item *entities.OrderItemProductDetail, userId string, branchId string) error {
	for _, itemStock := range item.Stocks {
		if itemStock.StockId != "" {
			stockID, err := primitive.ObjectIDFromHex(itemStock.StockId)
			if err != nil {
				return err
			}
			if _, err = entity.productStockRepo.UpdateOne(ctx, bson.M{"_id": stockID}, bson.M{
				"$inc": bson.M{"quantity": itemStock.Quantity},
			}); err != nil {
				return err
			}
		} else {
			if _, err := entity.productsRepo.UpdateOne(ctx, bson.M{"_id": item.ProductId}, bson.M{
				"$inc": bson.M{"soldFirst": itemStock.Quantity},
			}); err != nil {
				return err
			}
		}
	}

	unit := entities.ProductUnit{}
	if err := entity.productUnitsRepo.FindOne(ctx, bson.M{"_id": item.UnitId}).Decode(&unit); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	balance, err := entity.getProductStockBalanceWithContext(ctx, item.ProductId, unit.Id, branchId)
	if err != nil {
		return err
	}
	h := request.RemoveOrderItemProductHistory(item.ProductId.Hex(), unit.Unit, item, balance, userId)
	branchObjId, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return err
	}
	history := entities.ProductHistory{
		Id:          primitive.NewObjectID(),
		BranchId:    branchObjId,
		ProductId:   item.ProductId,
		Type:        h.Type,
		Description: h.Description,
		Unit:        h.Unit,
		Import:      h.Import,
		Quantity:    h.Quantity,
		CostPrice:   h.CostPrice,
		Price:       h.Price,
		Balance:     h.Balance,
		CreatedBy:   h.CreatedBy,
		CreatedDate: time.Now(),
	}
	_, err = entity.productHistoryRepo.InsertOne(ctx, history)
	return err
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
		{"$match": bson.M{"orderId": orderObjId, "productId": productObjId}},
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
	total, totalCost, err := entity.getOrderTotalsWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}
	var data entities.Order
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"total":       total,
		"totalCost":   totalCost,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) getOrderTotalsWithContext(ctx context.Context, orderId string) (float64, float64, error) {
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return 0, 0, err
	}
	pipeline := []bson.M{
		{"$match": bson.M{"orderId": objId}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$price"}, "totalCost": bson.M{"$sum": "$costPrice"}}},
	}
	var result []bson.M
	cursor, err := entity.orderItemRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	if err = cursor.All(ctx, &result); err != nil || len(result) == 0 {
		if err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	}
	var total float64
	var totalCost float64
	if v, ok := result[0]["total"].(float64); ok {
		total = v
	}
	if v, ok := result[0]["totalCost"].(float64); ok {
		totalCost = v
	}
	return total, totalCost, nil
}

func (entity *orderEntity) getProductStockBalanceWithContext(ctx context.Context, productId primitive.ObjectID, unitId primitive.ObjectID, branchId string) (int, error) {
	match := bson.M{"productId": productId, "unitId": unitId}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return 0, err
		}
		match["branchId"] = branchObjId
	}
	cursor, err := entity.productStockRepo.Aggregate(ctx, []bson.M{
		{"$match": match},
		{"$group": bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}},
	})
	if err != nil {
		return 0, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	switch v := results[0]["balance"].(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, nil
	}
}

func (entity *orderEntity) GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error) {
	logrus.Info("GetOrderSummary")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchFilter, err := buildActiveOrderAnalyticsMatchFilter(form.StartDate, form.EndDate, form.BranchId)
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

func (entity *orderEntity) GetOrderDailyChart(form request.GetOrderRange) ([]entities.OrderDailyChart, error) {
	logrus.Info("GetOrderDailyChart")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchFilter, err := buildActiveOrderAnalyticsMatchFilter(form.StartDate, form.EndDate, form.BranchId)
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

func (entity *orderEntity) GetOrderMonthlyChart(branchId string) ([]entities.OrderDailyChart, error) {
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
		"status":      constant.ACTIVE,
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
			"$gt": startDate,
			"$lt": endDate,
		},
		"status": constant.ACTIVE,
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

func (entity *orderEntity) GetABCAnalysis(branchId string) ([]entities.ABCProduct, error) {
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
	}
	orderMatch := bson.A{
		bson.M{"$eq": bson.A{"$_id", "$$oid"}},
		bson.M{"$eq": bson.A{"$status", constant.ACTIVE}},
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
			"totalRevenue": bson.M{"$sum": bson.M{"$multiply": bson.A{"$price", "$quantity"}}},
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
