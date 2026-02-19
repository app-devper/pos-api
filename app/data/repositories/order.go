package repositories

import (
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
	orderRepo     *mongo.Collection
	orderItemRepo *mongo.Collection
	paymentRepo   *mongo.Collection
}

type IOrder interface {
	CreateOrder(form request.Order) (*entities.Order, error)
	GetOrderRange(form request.GetOrderRange) ([]entities.Order, error)
	GetOrdersByCustomerCode(customerCode string) ([]entities.Order, error)
	UpdateTotal() ([]entities.Order, error)
	GetOrderById(id string) (*entities.Order, error)
	GetOrderDetailById(id string) (*entities.OrderDetail, error)
	UpdateTotalCostOrderById(id string, totalCost float64) (*entities.Order, error)
	UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error)
	RemoveOrderById(id string) (*entities.OrderDetail, error)
	UpdateTotalOrderById(id string) (*entities.Order, error)
	GetTotalOrderById(id string) float64
	GetTotalCostOrderById(id string) float64

	GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error)
	GetOrderItemById(id string) (*entities.OrderItem, error)
	UpdateOrderItemById(id string, form request.OrderItem) (*entities.OrderItem, error)
	RemoveOrderItemById(id string) (*entities.OrderItemProductDetail, error)
	GetOrderItemDetailById(id string) (*entities.OrderItemProductDetail, error)
	GetOrderItemDetailByOrderId(orderId string) ([]entities.OrderItemProductDetail, error)
	GetOrderItemDetailByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error)
	RemoveOrderItemByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error)
	GetOrderItemByProductId(productId string) ([]entities.OrderItem, error)
	GetOrderItemOrderDetailsByProductId(productId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error)

	GetPaymentByOrderId(orderId string) (*entities.Payment, error)
	RemovePaymentByOrderId(orderId string) (*entities.Payment, error)

	GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error)
	GetOrderDailyChart(form request.GetOrderRange) ([]entities.OrderDailyChart, error)
}

func NewOrderEntity(resource *db.Resource) IOrder {
	orderRepo := resource.PosDb.Collection("orders")
	orderItemRepo := resource.PosDb.Collection("order_items")
	paymentRepo := resource.PosDb.Collection("payments")
	entity := &orderEntity{orderRepo: orderRepo, orderItemRepo: orderItemRepo, paymentRepo: paymentRepo}
	ensureOrderIndexes(orderRepo, orderItemRepo, paymentRepo)
	return entity
}

func ensureOrderIndexes(orderRepo *mongo.Collection, orderItemRepo *mongo.Collection, paymentRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := orderRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create orders createdDate index: ", err)
	}

	_, err = orderRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "customerCode", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create orders customerCode index: ", err)
	}

	_, err = orderItemRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "orderId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create order_items orderId index: ", err)
	}

	_, err = orderItemRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create order_items productId index: ", err)
	}

	_, err = orderItemRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create order_items createdDate index: ", err)
	}

	_, err = paymentRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "orderId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create payments orderId index: ", err)
	}

	_, err = orderRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create orders branchId+createdDate index: ", err)
	}

	_, err = orderItemRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create order_items branchId+createdDate index: ", err)
	}

	_, err = paymentRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "orderId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create payments branchId+orderId index: ", err)
	}
}

func (entity *orderEntity) CreateOrder(form request.Order) (*entities.Order, error) {
	logrus.Info("CreateOrder")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	var orderId = primitive.NewObjectID()
	data := entities.Order{
		Id:           orderId,
		BranchId:     branchId,
		Code:         form.Code,
		CustomerCode: form.CustomerCode,
		CustomerName: form.CustomerName,
		Status:       constant.ACTIVE,
		Total:        form.Total,
		TotalCost:    form.TotalCost,
		Discount:     form.Discount,
		Type:         form.Type,
		CreatedBy:    form.CreatedBy,
		CreatedDate:  time.Now(),
		UpdatedBy:    form.CreatedBy,
		UpdatedDate:  time.Now(),
	}
	_, err := entity.orderRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}

	count := len(form.Items)
	orderItem := make([]interface{}, count)
	for i := 0; i < count; i++ {
		formItem := form.Items[i]
		productId, _ := primitive.ObjectIDFromHex(formItem.ProductId)
		unitId, _ := primitive.ObjectIDFromHex(formItem.UnitId)
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
		"$gt": form.StartDate,
		"$lt": form.EndDate,
	}}
	if form.BranchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(form.BranchId)
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
	return items, nil
}

func (entity *orderEntity) GetOrdersByCustomerCode(customerCode string) ([]entities.Order, error) {
	logrus.Info("GetOrdersByCustomerCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	var items []entities.Order
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})
	cursor, err := entity.orderRepo.Find(ctx, bson.M{"customerCode": customerCode}, opts)
	if err != nil {
		return nil, err
	}
	items = []entities.Order{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
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
			logrus.Error(err)
			logrus.Info(cursor.Current)
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

	payment, err := entity.GetPaymentByOrderId(id)
	if err != nil {
		return nil, err
	}
	data.Payment = *payment

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

	payment, _ := entity.RemovePaymentByOrderId(id)
	data.Payment = *payment

	items, _ := entity.RemoveOrderItemByOrderId(id)
	data.Items = items

	return &data, nil
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
	objId, _ := primitive.ObjectIDFromHex(orderId)
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
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		},
	}
	if form.BranchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(form.BranchId)
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

func (entity *orderEntity) GetPaymentByOrderId(orderId string) (*entities.Payment, error) {
	logrus.Info("GetPaymentByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	var data entities.Payment
	err = entity.paymentRepo.FindOne(ctx, bson.M{"orderId": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) RemovePaymentByOrderId(orderId string) (*entities.Payment, error) {
	logrus.Info("RemovePaymentByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	var data entities.Payment
	err = entity.paymentRepo.FindOneAndDelete(ctx, bson.M{"orderId": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) GetOrderSummary(form request.GetOrderRange) (*entities.OrderSummary, error) {
	logrus.Info("GetOrderSummary")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchFilter := bson.M{
		"createdDate": bson.M{
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		},
	}
	if form.BranchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(form.BranchId)
		matchFilter["branchId"] = branchObjId
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
			"totalProfit": bson.M{"$subtract": []string{"$totalRevenue", "$totalCost"}},
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

	matchFilter := bson.M{
		"createdDate": bson.M{
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		},
	}
	if form.BranchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(form.BranchId)
		matchFilter["branchId"] = branchObjId
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
			"totalProfit": bson.M{"$subtract": []string{"$totalRevenue", "$totalCost"}},
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
