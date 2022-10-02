package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/model"
	"pos/app/featues/request"
	"pos/db"
	"strings"
	"time"
)

type categoryEntity struct {
	categoryRepo *mongo.Collection
}

type ICategory interface {
	CreateIndex() (string, error)
	GetCategoryAll() ([]model.Category, error)
	CreateCategory(form request.Category) (*model.Category, error)
	GetCategoryById(id string) (*model.Category, error)
	RemoveCategoryById(id string) (*model.Category, error)
	UpdateCategoryById(id string, form request.Category) (*model.Category, error)
	UpdateDefaultCategoryById(id string) (*model.Category, error)
}

func NewCategoryEntity(resource *db.Resource) ICategory {
	categoryRepo := resource.PosDb.Collection("categories")
	var entity ICategory = &categoryEntity{categoryRepo: categoryRepo}
	_, _ = entity.CreateIndex()
	return entity
}

func (entity *categoryEntity) UpdateDefaultCategoryById(id string) (*model.Category, error) {
	logrus.Info("UpdateDefaultCategoryById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := entity.categoryRepo.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{
		"default": false,
	}})
	if err != nil {
		return nil, err
	}

	objId, _ := primitive.ObjectIDFromHex(id)
	var data model.Category
	err = entity.categoryRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Default = true
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.categoryRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *categoryEntity) GetCategoryAll() ([]model.Category, error) {
	logrus.Info("GetCategoryAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var items []model.Category
	cursor, err := entity.categoryRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var category model.Category
		err = cursor.Decode(&category)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, category)
		}
	}
	if items == nil {
		items = []model.Category{}
	}
	return items, nil
}

func (entity *categoryEntity) CreateCategory(form request.Category) (*model.Category, error) {
	logrus.Info("CreateCategory")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.Category{
		Id:          primitive.NewObjectID(),
		Name:        form.Name,
		Value:       strings.ToUpper(form.Value),
		Description: form.Description,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}
	_, err := entity.categoryRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *categoryEntity) GetCategoryById(id string) (*model.Category, error) {
	logrus.Info("GetCategoryById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data model.Category
	err := entity.categoryRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *categoryEntity) RemoveCategoryById(id string) (*model.Category, error) {
	logrus.Info("RemoveCategoryById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Category
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.categoryRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.categoryRepo.DeleteOne(ctx, bson.M{"_id": objId})
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *categoryEntity) UpdateCategoryById(id string, form request.Category) (*model.Category, error) {
	logrus.Info("UpdateCategoryById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	var data model.Category
	err := entity.categoryRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Name = form.Name
	data.Value = strings.ToUpper(form.Value)
	data.Description = form.Description
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.categoryRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *categoryEntity) CreateIndex() (string, error) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	mod := mongo.IndexModel{
		Keys: bson.M{
			"value": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	ind, err := entity.categoryRepo.Indexes().CreateOne(ctx, mod)
	return ind, err
}
