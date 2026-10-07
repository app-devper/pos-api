package repositories

import (
	"context"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/request"
	"pos/db"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type stockTransferEntity struct {
	ledger *ledger.Ledger
	repo   *mongo.Collection
}

type IStockTransfer interface {
	CreateStockTransferWithReservation(form request.StockTransfer) (*entities.StockTransfer, error)
	GetStockTransfers(branchId string) ([]entities.StockTransfer, error)
	GetStockTransferById(id string) (*entities.StockTransfer, error)
	ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error)
	RejectStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error)
}

func NewStockTransferEntity(resource *db.Resource) IStockTransfer {
	repo := resource.PosDb.Collection("stock_transfers")
	entity := &stockTransferEntity{ledger: newLedger(resource), repo: repo}
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

// CreateStockTransferWithReservation is recorded by the Stock ledger: the
// source quantity is reserved, with history, as the Transfer is created.
func (entity *stockTransferEntity) CreateStockTransferWithReservation(form request.StockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("CreateStockTransferWithReservation")
	return entity.ledger.RequestTransfer(context.Background(), form)
}

// ApproveStockTransfer is recorded by the Stock ledger: Stock opens at the
// destination and serves its waiting Lines; a second approve is refused.
func (entity *stockTransferEntity) ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	logrus.Info("ApproveStockTransfer")
	return entity.ledger.ApproveTransfer(context.Background(), id, updatedBy)
}

// RejectStockTransfer is recorded by the Stock ledger: the reservation goes
// back to its source Stocks.
func (entity *stockTransferEntity) RejectStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	logrus.Info("RejectStockTransfer")
	return entity.ledger.RejectTransfer(context.Background(), id, updatedBy)
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
