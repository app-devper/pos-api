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

type deliveryOrderEntity struct {
	repo *mongo.Collection
}

type IDeliveryOrder interface {
	CreateDeliveryOrder(form request.DeliveryOrder) (*entities.DeliveryOrder, error)
	GetDeliveryOrders(branchId string) ([]entities.DeliveryOrder, error)
	GetDeliveryOrderById(id string) (*entities.DeliveryOrder, error)
	UpdateDeliveryOrderById(id string, form request.UpdateDeliveryOrder) (*entities.DeliveryOrder, error)
	RemoveDeliveryOrderById(id string) (*entities.DeliveryOrder, error)
}

func NewDeliveryOrderEntity(resource *db.Resource) IDeliveryOrder {
	repo := resource.PosDb.Collection("delivery_orders")
	entity := &deliveryOrderEntity{repo: repo}
	ensureDeliveryOrderIndexes(repo)
	return entity
}

func ensureDeliveryOrderIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create delivery_orders index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create delivery_orders code index: ", err)
	}
}

func (entity *deliveryOrderEntity) CreateDeliveryOrder(form request.DeliveryOrder) (*entities.DeliveryOrder, error) {
	logrus.Info("CreateDeliveryOrder")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	orderId, _ := primitive.ObjectIDFromHex(form.OrderId)

	items := make([]entities.DeliveryOrderItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.DeliveryOrderItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	data := entities.DeliveryOrder{
		Id:           primitive.NewObjectID(),
		BranchId:     branchId,
		OrderId:      orderId,
		Code:         form.Code,
		CustomerCode: form.CustomerCode,
		CustomerName: form.CustomerName,
		Address:      form.Address,
		Items:        items,
		Note:         form.Note,
		Status:       constant.ACTIVE,
		DeliveryDate: form.DeliveryDate,
		CreatedBy:    form.CreatedBy,
		CreatedDate:  time.Now(),
		UpdatedBy:    form.CreatedBy,
		UpdatedDate:  time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *deliveryOrderEntity) GetDeliveryOrders(branchId string) ([]entities.DeliveryOrder, error) {
	logrus.Info("GetDeliveryOrders")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{}
	if branchId != "" {
		objId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = objId
	}

	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.DeliveryOrder
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DeliveryOrder{}
	}
	return results, nil
}

func (entity *deliveryOrderEntity) GetDeliveryOrderById(id string) (*entities.DeliveryOrder, error) {
	logrus.Info("GetDeliveryOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.DeliveryOrder{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *deliveryOrderEntity) UpdateDeliveryOrderById(id string, form request.UpdateDeliveryOrder) (*entities.DeliveryOrder, error) {
	logrus.Info("UpdateDeliveryOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	items := make([]entities.DeliveryOrderItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.DeliveryOrderItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"address":      form.Address,
		"items":        items,
		"note":         form.Note,
		"deliveryDate": form.DeliveryDate,
		"updatedBy":    form.UpdatedBy,
		"updatedDate":  time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.DeliveryOrder{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *deliveryOrderEntity) RemoveDeliveryOrderById(id string) (*entities.DeliveryOrder, error) {
	logrus.Info("RemoveDeliveryOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.DeliveryOrder{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
