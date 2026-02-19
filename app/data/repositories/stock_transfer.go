package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/request"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type stockTransferEntity struct {
	repo *mongo.Collection
}

type IStockTransfer interface {
	CreateStockTransfer(form request.StockTransfer) (*entities.StockTransfer, error)
	GetStockTransfers(branchId string) ([]entities.StockTransfer, error)
	GetStockTransferById(id string) (*entities.StockTransfer, error)
	UpdateStockTransferStatus(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error)
}

func NewStockTransferEntity(resource *db.Resource) IStockTransfer {
	repo := resource.PosDb.Collection("stock_transfers")
	entity := &stockTransferEntity{repo: repo}
	ensureStockTransferIndexes(repo)
	return entity
}

func ensureStockTransferIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "fromBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create stock_transfers fromBranch index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "toBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create stock_transfers toBranch index: ", err)
	}
}

func (entity *stockTransferEntity) CreateStockTransfer(form request.StockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("CreateStockTransfer")
	ctx, cancel := utils.InitContext()
	defer cancel()

	fromBranchId, _ := primitive.ObjectIDFromHex(form.FromBranchId)
	toBranchId, _ := primitive.ObjectIDFromHex(form.ToBranchId)

	items := make([]entities.StockTransferItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.StockTransferItem{
			ProductId: productId,
			StockId:   item.StockId,
			Quantity:  item.Quantity,
		}
	}

	data := entities.StockTransfer{
		Id:           primitive.NewObjectID(),
		FromBranchId: fromBranchId,
		ToBranchId:   toBranchId,
		Code:         form.Code,
		Items:        items,
		Note:         form.Note,
		Status:       "PENDING",
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

func (entity *stockTransferEntity) GetStockTransfers(branchId string) ([]entities.StockTransfer, error) {
	logrus.Info("GetStockTransfers")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(branchId)
	filter := bson.M{
		"$or": []bson.M{
			{"fromBranchId": objId},
			{"toBranchId": objId},
		},
	}
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.StockTransfer
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.StockTransfer{}
	}
	return results, nil
}

func (entity *stockTransferEntity) GetStockTransferById(id string) (*entities.StockTransfer, error) {
	logrus.Info("GetStockTransferById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.StockTransfer{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) UpdateStockTransferStatus(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("UpdateStockTransferStatus")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	data := entities.StockTransfer{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId, "status": "PENDING"}, bson.M{"$set": bson.M{
		"status":      form.Status,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
