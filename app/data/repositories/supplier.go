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
	ensureSupplierIndexes(supplierRepo)
	return entity
}

func ensureSupplierIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "clientId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create suppliers clientId index: ", err)
	}
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
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Supplier{}
	err = entity.supplierRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
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
	err = entity.supplierRepo.FindOneAndDelete(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
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
	items = []entities.Supplier{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *supplierEntity) CreateOrUpdateSupplierByClientId(id string, form request.Supplier) (*entities.Supplier, error) {
	logrus.Info("CreateOrUpdateSupplierByClientId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	now := time.Now()
	isReturnNewDoc := options.After
	upsert := true
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
		Upsert:         &upsert,
	}
	data := entities.Supplier{}
	err := entity.supplierRepo.FindOneAndUpdate(ctx, bson.M{"clientId": id}, bson.M{
		"$set": bson.M{
			"name":        form.Name,
			"address":     form.Address,
			"phone":       form.Phone,
			"taxId":       form.TaxId,
			"updatedBy":   form.UpdatedBy,
			"updatedDate": now,
		},
		"$setOnInsert": bson.M{
			"_id":         primitive.NewObjectID(),
			"clientId":    id,
			"createdBy":   form.UpdatedBy,
			"createdDate": now,
		},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
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

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	data := entities.Supplier{}
	err = entity.supplierRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": bson.M{
		"name":        form.Name,
		"address":     form.Address,
		"phone":       form.Phone,
		"taxId":       form.TaxId,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
