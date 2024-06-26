package repositories

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"time"
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
}

func NewOrderEntity(resource *db.Resource) IOrder {
	orderRepo := resource.PosDb.Collection("orders")
	orderItemRepo := resource.PosDb.Collection("order_items")
	paymentRepo := resource.PosDb.Collection("payments")
	entity := &orderEntity{orderRepo: orderRepo, orderItemRepo: orderItemRepo, paymentRepo: paymentRepo}
	return entity
}

func (entity *orderEntity) CreateOrder(form request.Order) (*entities.Order, error) {
	logrus.Info("CreateOrder")
	ctx, cancel := utils.InitContext()
	defer cancel()

	var orderId = primitive.NewObjectID()
	data := entities.Order{
		Id:           orderId,
		Code:         form.Code,
		CustomerCode: form.CustomerCode,
		CustomerName: form.CustomerName,
		Status:       constant.ACTIVE,
		Total:        form.Total,
		TotalCost:    form.TotalCost,
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

	payment := entities.Payment{
		Id:          primitive.NewObjectID(),
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
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	logrus.Info("GetOrderRange")
	ctx, cancel := utils.InitContext()
	defer cancel()

	var items []entities.Order
	cursor, err := entity.orderRepo.Find(ctx, bson.M{"createdDate": bson.M{
		"$gt": form.StartDate,
		"$lt": form.EndDate,
	},
	})
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
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.Order{}
	}
	return items, nil
}

func (entity *orderEntity) GetOrdersByCustomerCode(customerCode string) ([]entities.Order, error) {
	logrus.Info("GetOrdersByCustomerCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	var items []entities.Order
	cursor, err := entity.orderRepo.Find(ctx, bson.M{"customerCode": customerCode})
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
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.Order{}
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
	var data entities.Order
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) UpdateTotalCostOrderById(id string, totalCost float64) (*entities.Order, error) {
	logrus.Info("UpdateTotalCostOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data, err := entity.GetOrderById(id)
	objId, _ := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data.TotalCost = totalCost
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *orderEntity) UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error) {
	logrus.Info("UpdateCustomerCodeOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data, err := entity.GetOrderById(id)
	objId, _ := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data.CustomerCode = customerCode
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *orderEntity) GetOrderDetailById(id string) (*entities.OrderDetail, error) {
	logrus.Info("GetOrderDetailById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.OrderDetail
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
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

	var data entities.OrderDetail
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	_, err = entity.orderRepo.DeleteOne(ctx, bson.M{"_id": data.Id})
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

	var data entities.Order
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.orderRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	data.Total = entity.GetTotalOrderById(id)
	data.TotalCost = entity.GetTotalCostOrderById(id)
	data.UpdatedDate = time.Now()
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.orderRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (entity *orderEntity) GetTotalOrderById(orderId string) float64 {
	logrus.Info("GetTotalOrderById")
	ctx, cancel := utils.InitContext()
	objId, _ := primitive.ObjectIDFromHex(orderId)
	defer cancel()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"orderId": objId,
			},
		},
		{
			"$group": bson.M{
				"_id":   "",
				"total": bson.M{"$sum": "$price"},
			},
		},
	}
	var result []bson.M
	cursor, err := entity.orderItemRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	err = cursor.All(ctx, &result)
	if result == nil {
		return 0
	}
	return result[0]["total"].(float64)
}

func (entity *orderEntity) GetTotalCostOrderById(orderId string) float64 {
	logrus.Info("GetTotalCostOrderById")
	ctx, cancel := utils.InitContext()
	objId, _ := primitive.ObjectIDFromHex(orderId)
	defer cancel()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"orderId": objId,
			},
		},
		{
			"$group": bson.M{
				"_id":       "",
				"totalCost": bson.M{"$sum": "$costPrice"},
			},
		},
	}
	var result []bson.M
	cursor, err := entity.orderItemRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	err = cursor.All(ctx, &result)
	if result == nil {
		return 0
	}
	return result[0]["totalCost"].(float64)
}

func (entity *orderEntity) GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemRange")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cursor, err := entity.orderItemRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"createdDate": bson.M{
					"$gt": form.StartDate,
					"$lt": form.EndDate,
				},
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
	var items []entities.OrderItemProductDetail
	for cursor.Next(ctx) {
		var data entities.OrderItemProductDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItemProductDetail{}
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemById(id string) (*entities.OrderItem, error) {
	logrus.Info("GetOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.OrderItem
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.orderItemRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) UpdateOrderItemById(id string, form request.OrderItem) (*entities.OrderItem, error) {
	logrus.Info("UpdateOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data, err := entity.GetOrderItemById(id)
	objId, _ := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data.Discount = form.Discount
	data.Price = form.Price
	data.CostPrice = form.CostPrice
	data.Quantity = form.Quantity
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.orderItemRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *orderEntity) RemoveOrderItemById(id string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("RemoveOrderItemById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
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
	objId, _ := primitive.ObjectIDFromHex(id)
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
	var items []entities.OrderItemProductDetail
	for cursor.Next(ctx) {
		var data entities.OrderItemProductDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItemProductDetail{}
	}
	return &items[0], nil
}

func (entity *orderEntity) GetOrderItemDetailByOrderId(orderId string) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(orderId)
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
	var items []entities.OrderItemProductDetail
	for cursor.Next(ctx) {
		var data entities.OrderItemProductDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItemProductDetail{}
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemDetailByOrderProductId(orderId string, productId string) (*entities.OrderItemProductDetail, error) {
	logrus.Info("GetOrderItemDetailByOrderProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(orderId)
	productObjId, _ := primitive.ObjectIDFromHex(productId)
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
	var items []entities.OrderItemProductDetail
	for cursor.Next(ctx) {
		var data entities.OrderItemProductDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItemProductDetail{}
	}
	return &items[0], nil
}

func (entity *orderEntity) GetOrderItemByProductId(productId string) ([]entities.OrderItem, error) {
	logrus.Info("GetOrderItemByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.orderItemRepo.Find(ctx, bson.M{
		"productId": objId,
	})

	if err != nil {
		return nil, err
	}
	var items []entities.OrderItem
	for cursor.Next(ctx) {
		var data entities.OrderItem
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItem{}
	}
	return items, nil
}

func (entity *orderEntity) GetOrderItemOrderDetailsByProductId(productId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error) {
	logrus.Info("GetOrderItemOrderDetailsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	productObjId, _ := primitive.ObjectIDFromHex(productId)
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
	var items []entities.OrderItemOrderDetail
	for cursor.Next(ctx) {
		var data entities.OrderItemOrderDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.OrderItemOrderDetail{}
	}
	return items, nil
}

func (entity *orderEntity) RemoveOrderItemByOrderId(orderId string) ([]entities.OrderItemProductDetail, error) {
	logrus.Info("RemoveOrderItemByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(orderId)
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
	objId, _ := primitive.ObjectIDFromHex(orderId)
	productObjId, _ := primitive.ObjectIDFromHex(productId)
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
	var data entities.Payment
	objId, _ := primitive.ObjectIDFromHex(orderId)
	err := entity.paymentRepo.FindOne(ctx, bson.M{"orderId": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *orderEntity) RemovePaymentByOrderId(orderId string) (*entities.Payment, error) {
	logrus.Info("RemovePaymentByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.Payment
	objId, _ := primitive.ObjectIDFromHex(orderId)
	err := entity.paymentRepo.FindOne(ctx, bson.M{"orderId": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.paymentRepo.DeleteMany(ctx, bson.M{"orderId": objId})
	if err != nil {
		return nil, err
	}
	return &data, nil
}
