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

type purchaseOrderEntity struct {
	repo *mongo.Collection
}

type IPurchaseOrder interface {
	CreatePurchaseOrder(form request.PurchaseOrder) (*entities.PurchaseOrder, error)
	GetPurchaseOrders(branchId string) ([]entities.PurchaseOrder, error)
	GetPurchaseOrderById(id string) (*entities.PurchaseOrder, error)
	UpdatePurchaseOrderById(id string, form request.UpdatePurchaseOrder) (*entities.PurchaseOrder, error)
	RemovePurchaseOrderById(id string) (*entities.PurchaseOrder, error)
}

func NewPurchaseOrderEntity(resource *db.Resource) IPurchaseOrder {
	repo := resource.PosDb.Collection("purchase_orders")
	entity := &purchaseOrderEntity{repo: repo}
	ensurePurchaseOrderIndexes(repo)
	return entity
}

func ensurePurchaseOrderIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create purchase_orders index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create purchase_orders code index: ", err)
	}
}

func (entity *purchaseOrderEntity) CreatePurchaseOrder(form request.PurchaseOrder) (*entities.PurchaseOrder, error) {
	logrus.Info("CreatePurchaseOrder")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	supplierId, _ := primitive.ObjectIDFromHex(form.SupplierId)

	items := make([]entities.PurchaseOrderItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.PurchaseOrderItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			CostPrice: item.CostPrice,
			Total:     item.Total,
		}
	}

	data := entities.PurchaseOrder{
		Id:          primitive.NewObjectID(),
		BranchId:    branchId,
		SupplierId:  supplierId,
		Code:        form.Code,
		Reference:   form.Reference,
		Items:       items,
		TotalCost:   form.TotalCost,
		Note:        form.Note,
		Status:      constant.ACTIVE,
		DueDate:     form.DueDate,
		CreatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedBy:   form.CreatedBy,
		UpdatedDate: time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *purchaseOrderEntity) GetPurchaseOrders(branchId string) ([]entities.PurchaseOrder, error) {
	logrus.Info("GetPurchaseOrders")
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
	var results []entities.PurchaseOrder
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.PurchaseOrder{}
	}
	return results, nil
}

func (entity *purchaseOrderEntity) GetPurchaseOrderById(id string) (*entities.PurchaseOrder, error) {
	logrus.Info("GetPurchaseOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.PurchaseOrder{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *purchaseOrderEntity) UpdatePurchaseOrderById(id string, form request.UpdatePurchaseOrder) (*entities.PurchaseOrder, error) {
	logrus.Info("UpdatePurchaseOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	supplierId, _ := primitive.ObjectIDFromHex(form.SupplierId)

	items := make([]entities.PurchaseOrderItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.PurchaseOrderItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			CostPrice: item.CostPrice,
			Total:     item.Total,
		}
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"supplierId":  supplierId,
		"reference":   form.Reference,
		"items":       items,
		"totalCost":   form.TotalCost,
		"note":        form.Note,
		"dueDate":     form.DueDate,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.PurchaseOrder{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *purchaseOrderEntity) RemovePurchaseOrderById(id string) (*entities.PurchaseOrder, error) {
	logrus.Info("RemovePurchaseOrderById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.PurchaseOrder{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
