package repositories

import (
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

type productReturnEntity struct {
	ledger            *ledger.Ledger
	productReturnRepo *mongo.Collection
}

type IProductReturn interface {
	RecordProductReturn(req request.ProductReturn) (*entities.ProductReturn, error)
	GetProductReturnById(id string) (*entities.ProductReturn, error)
	GetProductReturnsByOrderId(orderId string, branchId string) ([]entities.ProductReturn, error)
}

func NewProductReturnEntity(resource *db.Resource) IProductReturn {
	productReturnRepo := resource.PosDb.Collection("product_returns")
	entity := &productReturnEntity{ledger: newLedger(resource), productReturnRepo: productReturnRepo}
	ensureProductReturnIndexes(productReturnRepo)
	return entity
}

func ensureProductReturnIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "product_returns orderId", mongo.IndexModel{
		Keys: bson.D{{Key: "orderId", Value: 1}},
	})
	createCollectionIndex(repo, "product_returns branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *productReturnEntity) GetProductReturnById(id string) (*entities.ProductReturn, error) {
	logrus.Info("GetProductReturnById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.ProductReturn{}
	if err := entity.productReturnRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productReturnEntity) GetProductReturnsByOrderId(orderId string, branchId string) ([]entities.ProductReturn, error) {
	logrus.Info("GetProductReturnsByOrderId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"orderId": objId}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})
	cursor, err := entity.productReturnRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items := []entities.ProductReturn{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
