package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"time"
)

type customerEntity struct {
	customerRepo *mongo.Collection
}

type ICustomer interface {
	CreateIndex() (string, error)
	GetCustomerAll() ([]entities.Customer, error)
	CreateCustomer(form request.Customer) (*entities.Customer, error)
	GetCustomerById(id string) (*entities.Customer, error)
	GetCustomerByCode(code string) (*entities.Customer, error)
	RemoveCustomerById(id string) (*entities.Customer, error)
	UpdateCustomerById(id string, form request.UpdateCustomer) (*entities.Customer, error)
	UpdateCustomerStatusById(id string, form request.UpdateCustomerStatus) (*entities.Customer, error)
}

func NewCustomerEntity(resource *db.Resource) ICustomer {
	customerRepo := resource.PosDb.Collection("customers")
	entity := &customerEntity{customerRepo: customerRepo}
	_, _ = entity.CreateIndex()
	return entity
}

func (entity *customerEntity) GetCustomerAll() ([]entities.Customer, error) {
	logrus.Info("GetCustomerAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var items []entities.Customer
	cursor, err := entity.customerRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var item entities.Customer
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.Customer{}
	}
	return items, nil
}

func (entity *customerEntity) CreateCustomer(form request.Customer) (*entities.Customer, error) {
	logrus.Info("CreateCustomer")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Customer{
		Id:           primitive.NewObjectID(),
		CustomerType: form.CustomerType,
		Code:         form.Code,
		Name:         form.Name,
		Address:      form.Address,
		Phone:        form.Phone,
		Email:        form.Email,
		Status:       constant.ACTIVE,
		CreatedBy:    form.CreatedBy,
		UpdatedBy:    form.CreatedBy,
		CreatedDate:  time.Now(),
		UpdatedDate:  time.Now(),
	}
	_, err := entity.customerRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) GetCustomerById(id string) (*entities.Customer, error) {
	logrus.Info("GetCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data entities.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) GetCustomerByCode(code string) (*entities.Customer, error) {
	logrus.Info("GetCustomerByCode")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) RemoveCustomerById(id string) (*entities.Customer, error) {
	logrus.Info("RemoveCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.Customer
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.customerRepo.DeleteOne(ctx, bson.M{"_id": objId})
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) UpdateCustomerById(id string, form request.UpdateCustomer) (*entities.Customer, error) {
	logrus.Info("UpdateCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data entities.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	data.CustomerType = form.CustomerType
	data.Name = form.Name
	data.Address = form.Address
	data.Phone = form.Phone
	data.Email = form.Email
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.customerRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) UpdateCustomerStatusById(id string, form request.UpdateCustomerStatus) (*entities.Customer, error) {
	logrus.Info("UpdateCustomerStatusById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cid, _ := primitive.ObjectIDFromHex(id)
	var data entities.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": cid}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Status = form.Status
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.customerRepo.FindOneAndUpdate(ctx, bson.M{"_id": cid}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) CreateIndex() (string, error) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	mod := mongo.IndexModel{
		Keys: bson.M{
			"code": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	ind, err := entity.customerRepo.Indexes().CreateOne(ctx, mod)
	return ind, err
}
