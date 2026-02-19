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

type branchEntity struct {
	branchRepo *mongo.Collection
}

type IBranch interface {
	CreateBranch(form request.Branch) (*entities.Branch, error)
	GetBranches() ([]entities.Branch, error)
	GetBranchById(id string) (*entities.Branch, error)
	GetBranchByCode(code string) (*entities.Branch, error)
	UpdateBranchById(id string, form request.UpdateBranch) (*entities.Branch, error)
	UpdateBranchStatusById(id string, form request.UpdateBranchStatus) (*entities.Branch, error)
	RemoveBranchById(id string) (*entities.Branch, error)
}

func NewBranchEntity(resource *db.Resource) IBranch {
	branchRepo := resource.PosDb.Collection("branches")
	entity := &branchEntity{branchRepo: branchRepo}
	ensureBranchIndexes(branchRepo)
	return entity
}

func ensureBranchIndexes(branchRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := branchRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create branches code index: ", err)
	}
}

func (entity *branchEntity) CreateBranch(form request.Branch) (*entities.Branch, error) {
	logrus.Info("CreateBranch")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := entities.Branch{
		Id:          primitive.NewObjectID(),
		Code:        form.Code,
		Name:        form.Name,
		Address:     form.Address,
		Phone:       form.Phone,
		Status:      constant.ACTIVE,
		CreatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedBy:   form.CreatedBy,
		UpdatedDate: time.Now(),
	}

	_, err := entity.branchRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *branchEntity) GetBranches() (items []entities.Branch, err error) {
	logrus.Info("GetBranches")
	ctx, cancel := utils.InitContext()
	defer cancel()

	cursor, err := entity.branchRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []entities.Branch{}
	}
	return items, nil
}

func (entity *branchEntity) GetBranchById(id string) (*entities.Branch, error) {
	logrus.Info("GetBranchById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data := entities.Branch{}
	err = entity.branchRepo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *branchEntity) GetBranchByCode(code string) (*entities.Branch, error) {
	logrus.Info("GetBranchByCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := entities.Branch{}
	err := entity.branchRepo.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *branchEntity) UpdateBranchById(id string, form request.UpdateBranch) (*entities.Branch, error) {
	logrus.Info("UpdateBranchById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}

	data := entities.Branch{}
	err = entity.branchRepo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": bson.M{
		"name":        form.Name,
		"address":     form.Address,
		"phone":       form.Phone,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *branchEntity) UpdateBranchStatusById(id string, form request.UpdateBranchStatus) (*entities.Branch, error) {
	logrus.Info("UpdateBranchStatusById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}

	data := entities.Branch{}
	err = entity.branchRepo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": bson.M{
		"status":      form.Status,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *branchEntity) RemoveBranchById(id string) (*entities.Branch, error) {
	logrus.Info("RemoveBranchById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	data := entities.Branch{}
	err = entity.branchRepo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
