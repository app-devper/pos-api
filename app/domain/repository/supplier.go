package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/model"
	"pos/app/domain/request"
	"pos/db"
	"time"
)

type supplierEntity struct {
	supplierRepo *mongo.Collection
}

type ISupplier interface {
	GetSupplierByClientId(id string) (*model.Supplier, error)
	CreateSupplierByClientId(id string, form request.Supplier) (*model.Supplier, error)
}

func NewSupplierEntity(resource *db.Resource) ISupplier {
	supplierRepo := resource.PosDb.Collection("suppliers")
	var entity ISupplier = &supplierEntity{supplierRepo: supplierRepo}
	return entity
}

func (entity *supplierEntity) GetSupplierByClientId(id string) (*model.Supplier, error) {
	logrus.Info("GetSupplierByClientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Supplier
	err := entity.supplierRepo.FindOne(ctx, bson.M{"clientId": id}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *supplierEntity) CreateSupplierByClientId(id string, form request.Supplier) (*model.Supplier, error) {
	logrus.Info("CreateSupplierByClientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Supplier
	err := entity.supplierRepo.FindOne(ctx, bson.M{"clientId": id}).Decode(&data)
	if err == nil {
		data.Name = form.Name
		data.Address = form.Address
		data.Phone = form.Phone
		data.TaxId = form.TaxId
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()
		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err := entity.supplierRepo.FindOneAndUpdate(ctx, bson.M{"clientId": id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	} else {
		data := model.Supplier{}
		data.Id = primitive.NewObjectID()
		data.ClientId = id
		data.Name = form.Name
		data.Address = form.Address
		data.Phone = form.Phone
		data.TaxId = form.TaxId
		data.CreatedBy = form.CreatedBy
		data.CreatedDate = time.Now()
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()
		_, err := entity.supplierRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}
