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

type productReturnEntity struct {
	productReturnRepo *mongo.Collection
}

type IProductReturn interface {
	CreateProductReturn(param ProductReturnInput) (*entities.ProductReturn, error)
	GetProductReturnById(id string) (*entities.ProductReturn, error)
	GetProductReturnsByOrderId(orderId string, branchId string) ([]entities.ProductReturn, error)
}

type ProductReturnInput struct {
	BranchId     primitive.ObjectID
	ReturnNo     string
	OrderId      primitive.ObjectID
	CustomerCode string
	Reason       string
	Note         string
	Items        []entities.ProductReturnItem
	TotalRefund  float64
	CreatedBy    string
}

func NewProductReturnEntity(resource *db.Resource) IProductReturn {
	productReturnRepo := resource.PosDb.Collection("product_returns")
	entity := &productReturnEntity{productReturnRepo: productReturnRepo}
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

func (entity *productReturnEntity) CreateProductReturn(param ProductReturnInput) (*entities.ProductReturn, error) {
	logrus.Info("CreateProductReturn")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductReturn{
		Id:           primitive.NewObjectID(),
		BranchId:     param.BranchId,
		ReturnNo:     param.ReturnNo,
		OrderId:      param.OrderId,
		CustomerCode: param.CustomerCode,
		Reason:       param.Reason,
		Note:         param.Note,
		Items:        param.Items,
		TotalRefund:  param.TotalRefund,
		CreatedBy:    param.CreatedBy,
		CreatedDate:  time.Now(),
	}
	if _, err := entity.productReturnRepo.InsertOne(ctx, data); err != nil {
		return nil, err
	}
	return &data, nil
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
