package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/entities"
	"pos/app/domain/request"
	"pos/db"
	"time"
)

type supplierEntity struct {
	supplierRepo *mongo.Collection
}

type ISupplier interface {
	GetSupplierByClientId(id string) (*entities.Supplier, error)
	GetSupplierById(id string) (*entities.Supplier, error)
	RemoveSupplierById(id string) (*entities.Supplier, error)
	GetSuppliers() ([]entities.Supplier, error)
	CreateOrUpdateSupplierByClientId(id string, form request.Supplier) (*entities.Supplier, error)
	CreateSupplier(form request.Supplier) (*entities.Supplier, error)
	UpdateSupplierById(id string, form request.Supplier) (*entities.Supplier, error)
}

func NewSupplierEntity(resource *db.Resource) ISupplier {
	supplierRepo := resource.PosDb.Collection("suppliers")
	entity := &supplierEntity{supplierRepo: supplierRepo}
	return entity
}

func (entity *supplierEntity) GetSupplierByClientId(id string) (*entities.Supplier, error) {
	logrus.Info("GetSupplierByClientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Supplier{}
	err := entity.supplierRepo.FindOne(ctx, bson.M{"clientId": id}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *supplierEntity) GetSupplierById(id string) (*entities.Supplier, error) {
	logrus.Info("GetSupplierById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Supplier{}
	err := entity.supplierRepo.FindOne(ctx, bson.M{"_id": id}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *supplierEntity) RemoveSupplierById(id string) (*entities.Supplier, error) {
	logrus.Info("RemoveSupplierById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Supplier{}
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.supplierRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.supplierRepo.DeleteOne(ctx, bson.M{"_id": obId})
	if err != nil {
		logrus.Error(err)
	}
	return &data, nil
}

func (entity *supplierEntity) GetSuppliers() (items []entities.Supplier, err error) {
	logrus.Info("GetSuppliers")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cursor, err := entity.supplierRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := entities.Supplier{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.Supplier{}
	}
	return items, nil
}

func (entity *supplierEntity) CreateOrUpdateSupplierByClientId(id string, form request.Supplier) (*entities.Supplier, error) {
	logrus.Info("CreateOrUpdateSupplierByClientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Supplier{}
	err := entity.supplierRepo.FindOne(ctx, bson.M{"clientId": id}).Decode(&data)
	if err == nil {
		data.Name = form.Name
		data.Address = form.Address
		data.Phone = form.Phone
		data.TaxId = form.TaxId
		data.UpdatedBy = form.UpdatedBy
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
		data.Id = primitive.NewObjectID()
		data.ClientId = id
		data.Name = form.Name
		data.Address = form.Address
		data.Phone = form.Phone
		data.TaxId = form.TaxId
		data.CreatedBy = form.UpdatedBy
		data.CreatedDate = time.Now()
		data.UpdatedBy = form.UpdatedBy
		data.UpdatedDate = time.Now()
		_, err := entity.supplierRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}

func (entity *supplierEntity) CreateSupplier(form request.Supplier) (*entities.Supplier, error) {
	logrus.Info("CreateSupplier")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.Supplier{}
	data.Id = primitive.NewObjectID()
	data.Name = form.Name
	data.Address = form.Address
	data.Phone = form.Phone
	data.TaxId = form.TaxId
	data.CreatedBy = form.UpdatedBy
	data.CreatedDate = time.Now()
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()
	_, err := entity.supplierRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *supplierEntity) UpdateSupplierById(id string, form request.Supplier) (*entities.Supplier, error) {
	logrus.Info("UpdateSupplierById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Supplier{}
	err = entity.supplierRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	data.Name = form.Name
	data.Address = form.Address
	data.Phone = form.Phone
	data.TaxId = form.TaxId
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.supplierRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil

}
