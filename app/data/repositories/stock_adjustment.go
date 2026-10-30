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

type stockAdjustmentEntity struct {
	stockAdjustmentRepo *mongo.Collection
}

type IStockAdjustment interface {
	GetStockAdjustmentsByProductId(productId string, branchId string) ([]entities.StockAdjustment, error)
}

func NewStockAdjustmentEntity(resource *db.Resource) IStockAdjustment {
	stockAdjustmentRepo := resource.PosDb.Collection("stock_adjustments")
	entity := &stockAdjustmentEntity{stockAdjustmentRepo: stockAdjustmentRepo}
	ensureStockAdjustmentIndexes(stockAdjustmentRepo)
	return entity
}

func ensureStockAdjustmentIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "stock_adjustments productId", mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	createCollectionIndex(repo, "stock_adjustments branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *stockAdjustmentEntity) GetStockAdjustmentsByProductId(productId string, branchId string) ([]entities.StockAdjustment, error) {
	logrus.Info("GetStockAdjustmentsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"productId": objId}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})
	cursor, err := entity.stockAdjustmentRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items := []entities.StockAdjustment{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
