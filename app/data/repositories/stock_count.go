package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type stockCountEntity struct {
	stockCountRepo *mongo.Collection
}

type IStockCount interface {
	CreateStockCount(param StockCountInput) (*entities.StockCount, error)
	GetStockCountById(id string) (*entities.StockCount, error)
	GetStockCounts(branchId string) ([]entities.StockCount, error)
}

type StockCountInput struct {
	BranchId  primitive.ObjectID
	CountNo   string
	Note      string
	Items     []entities.StockCountItem
	CreatedBy string
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

func (entity *stockCountEntity) CreateStockCount(param StockCountInput) (*entities.StockCount, error) {
	logrus.Info("CreateStockCount")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.StockCount{
		Id:          primitive.NewObjectID(),
		BranchId:    param.BranchId,
		CountNo:     param.CountNo,
		Note:        param.Note,
		Items:       param.Items,
		CreatedBy:   param.CreatedBy,
		CreatedDate: time.Now(),
	}
	if _, err := entity.stockCountRepo.InsertOne(ctx, data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockCountEntity) GetStockCountById(id string) (*entities.StockCount, error) {
	logrus.Info("GetStockCountById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.StockCount{}
	if err := entity.stockCountRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockCountEntity) GetStockCounts(branchId string) ([]entities.StockCount, error) {
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
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})
	cursor, err := entity.stockCountRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items := []entities.StockCount{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
