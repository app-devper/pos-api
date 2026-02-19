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

type employeeEntity struct {
	employeeRepo *mongo.Collection
}

type IEmployee interface {
	CreateEmployee(form request.Employee) (*entities.Employee, error)
	GetEmployees() ([]entities.Employee, error)
	GetEmployeesByBranchId(branchId string) ([]entities.Employee, error)
	GetEmployeeById(id string) (*entities.Employee, error)
	GetEmployeeByUserId(userId string) (*entities.Employee, error)
	UpdateEmployeeById(id string, form request.UpdateEmployee) (*entities.Employee, error)
	RemoveEmployeeById(id string) (*entities.Employee, error)
}

func NewEmployeeEntity(resource *db.Resource) IEmployee {
	employeeRepo := resource.PosDb.Collection("employees")
	entity := &employeeEntity{employeeRepo: employeeRepo}
	ensureEmployeeIndexes(employeeRepo)
	return entity
}

func ensureEmployeeIndexes(employeeRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := employeeRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create employees branchId index: ", err)
	}

	_, err = employeeRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create employees userId index: ", err)
	}
}

func (entity *employeeEntity) CreateEmployee(form request.Employee) (*entities.Employee, error) {
	logrus.Info("CreateEmployee")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}

	data := entities.Employee{
		Id:          primitive.NewObjectID(),
		BranchId:    branchId,
		UserId:      form.UserId,
		Role:        form.Role,
		CreatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedBy:   form.CreatedBy,
		UpdatedDate: time.Now(),
	}

	_, err = entity.employeeRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *employeeEntity) GetEmployees() (items []entities.Employee, err error) {
	logrus.Info("GetEmployees")
	ctx, cancel := utils.InitContext()
	defer cancel()

	cursor, err := entity.employeeRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []entities.Employee{}
	}
	return items, nil
}

func (entity *employeeEntity) GetEmployeesByBranchId(branchId string) (items []entities.Employee, err error) {
	logrus.Info("GetEmployeesByBranchId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}

	cursor, err := entity.employeeRepo.Find(ctx, bson.M{"branchId": objectId})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []entities.Employee{}
	}
	return items, nil
}

func (entity *employeeEntity) GetEmployeeById(id string) (*entities.Employee, error) {
	logrus.Info("GetEmployeeById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data := entities.Employee{}
	err = entity.employeeRepo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *employeeEntity) GetEmployeeByUserId(userId string) (*entities.Employee, error) {
	logrus.Info("GetEmployeeByUserId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := entities.Employee{}
	err := entity.employeeRepo.FindOne(ctx, bson.M{"userId": userId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *employeeEntity) UpdateEmployeeById(id string, form request.UpdateEmployee) (*entities.Employee, error) {
	logrus.Info("UpdateEmployeeById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}

	data := entities.Employee{}
	err = entity.employeeRepo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": bson.M{
		"branchId":    branchId,
		"role":        form.Role,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *employeeEntity) RemoveEmployeeById(id string) (*entities.Employee, error) {
	logrus.Info("RemoveEmployeeById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data := entities.Employee{}
	err = entity.employeeRepo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
