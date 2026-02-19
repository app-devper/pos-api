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

type customerHistoryEntity struct {
	repo *mongo.Collection
}

type ICustomerHistory interface {
	CreateCustomerHistory(form request.CustomerHistory) (*entities.CustomerHistory, error)
	GetCustomerHistories(customerCode string, branchId string) ([]entities.CustomerHistory, error)
}

func NewCustomerHistoryEntity(resource *db.Resource) ICustomerHistory {
	repo := resource.PosDb.Collection("customer_histories")
	entity := &customerHistoryEntity{repo: repo}
	ensureCustomerHistoryIndexes(repo)
	return entity
}

func ensureCustomerHistoryIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "customerCode", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create customer_histories index: ", err)
	}
}

func (entity *customerHistoryEntity) CreateCustomerHistory(form request.CustomerHistory) (*entities.CustomerHistory, error) {
	logrus.Info("CreateCustomerHistory")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	data := entities.CustomerHistory{
		Id:           primitive.NewObjectID(),
		BranchId:     branchId,
		CustomerCode: form.CustomerCode,
		Type:         form.Type,
		Description:  form.Description,
		Reference:    form.Reference,
		CreatedBy:    form.CreatedBy,
		CreatedDate:  time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerHistoryEntity) GetCustomerHistories(customerCode string, branchId string) ([]entities.CustomerHistory, error) {
	logrus.Info("GetCustomerHistories")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{"customerCode": customerCode}
	if branchId != "" {
		objId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = objId
	}

	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.CustomerHistory
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.CustomerHistory{}
	}
	return results, nil
}
