package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/db"

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
	GetStockTransfers(branchId string) ([]entities.StockTransfer, error)
	GetStockTransferById(id string, branchId string) (*entities.StockTransfer, error)
}

func NewStockTransferEntity(resource *db.Resource) IStockTransfer {
	repo := resource.PosDb.Collection("stock_transfers")
	entity := &stockTransferEntity{repo: repo}
	ensureStockTransferIndexes(repo)
	return entity
}

func ensureStockTransferIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "stock_transfers fromBranchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "fromBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(repo, "stock_transfers toBranchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "toBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *stockTransferEntity) GetStockTransfers(branchId string) ([]entities.StockTransfer, error) {
	logrus.Info("GetStockTransfers")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objId, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}
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

// GetStockTransferById reads a Transfer that branchId sends or receives.
func (entity *stockTransferEntity) GetStockTransferById(id string, branchId string) (*entities.StockTransfer, error) {
	logrus.Info("GetStockTransferById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	branch, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}
	data := entities.StockTransfer{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId, "$or": bson.A{
		bson.M{"fromBranchId": branch}, bson.M{"toBranchId": branch}}}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
