package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/db"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type stockCountEntity struct {
	stockCountRepo *mongo.Collection
}

type IStockCount interface {
	GetStockCountById(id string, branchId string) (*entities.StockCountDetail, error)
	GetStockCounts(branchId string) ([]entities.StockCountDetail, error)
}

func NewStockCountEntity(resource *db.Resource) IStockCount {
	stockCountRepo := resource.PosDb.Collection("stock_counts")
	entity := &stockCountEntity{stockCountRepo: stockCountRepo}
	ensureStockCountIndexes(stockCountRepo)
	return entity
}

func ensureStockCountIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "stock_counts branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *stockCountEntity) GetStockCountById(id string, branchId string) (*entities.StockCountDetail, error) {
	logrus.Info("GetStockCountById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter, err := inBranch(id, branchId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.stockCountRepo.Aggregate(ctx, append([]bson.M{{"$match": filter}}, namedCountLines()...))
	if err != nil {
		return nil, err
	}
	items := []entities.StockCountDetail{}
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &items[0], nil
}

func (entity *stockCountEntity) GetStockCounts(branchId string) ([]entities.StockCountDetail, error) {
	logrus.Info("GetStockCounts")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter := bson.M{}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	pipeline := append([]bson.M{{"$match": filter}, {"$sort": bson.D{{Key: "createdDate", Value: -1}}}}, namedCountLines()...)
	cursor, err := entity.stockCountRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	items := []entities.StockCountDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
