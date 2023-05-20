package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/constant"
	"pos/app/domain/model"
	"pos/app/domain/request"
	"pos/db"
	"time"
)

type customerEntity struct {
	customerRepo *mongo.Collection
}

type ICustomer interface {
	CreateIndex() (string, error)
	GetCustomerAll() ([]model.Customer, error)
	CreateCustomer(form request.Customer) (*model.Customer, error)
	GetCustomerById(id string) (*model.Customer, error)
	GetCustomerByCode(code string) (*model.Customer, error)
	RemoveCustomerById(id string) (*model.Customer, error)
	UpdateCustomerById(id string, form request.UpdateCustomer) (*model.Customer, error)
	UpdateCustomerStatusById(id string, form request.UpdateCustomerStatus) (*model.Customer, error)
}

func NewCustomerEntity(resource *db.Resource) ICustomer {
	customerRepo := resource.PosDb.Collection("customers")
	var entity ICustomer = &customerEntity{customerRepo: customerRepo}
	_, _ = entity.CreateIndex()
	return entity
}

func (entity *customerEntity) GetCustomerAll() ([]model.Customer, error) {
	logrus.Info("GetCustomerAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var items []model.Customer
	cursor, err := entity.customerRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var item model.Customer
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.Customer{}
	}
	return items, nil
}

func (entity *customerEntity) CreateCustomer(form request.Customer) (*model.Customer, error) {
	logrus.Info("CreateCustomer")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.Customer{
		Id:          primitive.NewObjectID(),
		Code:        form.Code,
		Name:        form.Name,
		Address:     form.Address,
		Phone:       form.Phone,
		Email:       form.Email,
		Status:      constant.ACTIVE,
		CreatedBy:   form.CreatedBy,
		UpdatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}
	_, err := entity.customerRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) GetCustomerById(id string) (*model.Customer, error) {
	logrus.Info("GetCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data model.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) GetCustomerByCode(code string) (*model.Customer, error) {
	logrus.Info("GetCustomerByCode")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *customerEntity) RemoveCustomerById(id string) (*model.Customer, error) {
	logrus.Info("RemoveCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Customer
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

func (entity *customerEntity) UpdateCustomerById(id string, form request.UpdateCustomer) (*model.Customer, error) {
	logrus.Info("UpdateCustomerById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data model.Customer
	err := entity.customerRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
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

func (entity *customerEntity) UpdateCustomerStatusById(id string, form request.UpdateCustomerStatus) (*model.Customer, error) {
	logrus.Info("UpdateCustomerStatusById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cid, _ := primitive.ObjectIDFromHex(id)
	var data model.Customer
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
