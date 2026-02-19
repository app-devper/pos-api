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

type billingEntity struct {
	repo *mongo.Collection
}

type IBilling interface {
	CreateBilling(form request.Billing) (*entities.Billing, error)
	GetBillings(branchId string) ([]entities.Billing, error)
	GetBillingById(id string) (*entities.Billing, error)
	UpdateBillingById(id string, form request.UpdateBilling) (*entities.Billing, error)
	RemoveBillingById(id string) (*entities.Billing, error)
}

func NewBillingEntity(resource *db.Resource) IBilling {
	repo := resource.PosDb.Collection("billings")
	entity := &billingEntity{repo: repo}
	ensureBillingIndexes(repo)
	return entity
}

func ensureBillingIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create billings index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create billings code index: ", err)
	}
}

func (entity *billingEntity) CreateBilling(form request.Billing) (*entities.Billing, error) {
	logrus.Info("CreateBilling")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)

	orderIds := make([]primitive.ObjectID, len(form.OrderIds))
	for i, id := range form.OrderIds {
		orderIds[i], _ = primitive.ObjectIDFromHex(id)
	}

	data := entities.Billing{
		Id:           primitive.NewObjectID(),
		BranchId:     branchId,
		CustomerCode: form.CustomerCode,
		CustomerName: form.CustomerName,
		Code:         form.Code,
		OrderIds:     orderIds,
		TotalAmount:  form.TotalAmount,
		Note:         form.Note,
		Status:       constant.ACTIVE,
		DueDate:      form.DueDate,
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

func (entity *billingEntity) GetBillings(branchId string) ([]entities.Billing, error) {
	logrus.Info("GetBillings")
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
	var results []entities.Billing
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.Billing{}
	}
	return results, nil
}

func (entity *billingEntity) GetBillingById(id string) (*entities.Billing, error) {
	logrus.Info("GetBillingById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Billing{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *billingEntity) UpdateBillingById(id string, form request.UpdateBilling) (*entities.Billing, error) {
	logrus.Info("UpdateBillingById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	orderIds := make([]primitive.ObjectID, len(form.OrderIds))
	for i, oid := range form.OrderIds {
		orderIds[i], _ = primitive.ObjectIDFromHex(oid)
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"orderIds":    orderIds,
		"totalAmount": form.TotalAmount,
		"note":        form.Note,
		"dueDate":     form.DueDate,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.Billing{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *billingEntity) RemoveBillingById(id string) (*entities.Billing, error) {
	logrus.Info("RemoveBillingById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Billing{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
