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

type promotionEntity struct {
	repo *mongo.Collection
}

type IPromotion interface {
	CreatePromotion(form request.Promotion) (*entities.Promotion, error)
	GetPromotions(branchId string) ([]entities.Promotion, error)
	GetPromotionById(id string) (*entities.Promotion, error)
	GetPromotionByCode(code string, branchId string) (*entities.Promotion, error)
	UpdatePromotionById(id string, form request.UpdatePromotion) (*entities.Promotion, error)
	RemovePromotionById(id string) (*entities.Promotion, error)
}

func NewPromotionEntity(resource *db.Resource) IPromotion {
	repo := resource.PosDb.Collection("promotions")
	entity := &promotionEntity{repo: repo}
	ensurePromotionIndexes(repo)
	return entity
}

func ensurePromotionIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "status", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create promotions index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}, {Key: "branchId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create promotions code index: ", err)
	}
}

func (entity *promotionEntity) CreatePromotion(form request.Promotion) (*entities.Promotion, error) {
	logrus.Info("CreatePromotion")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	productIds := make([]primitive.ObjectID, len(form.ProductIds))
	for i, id := range form.ProductIds {
		productIds[i], _ = primitive.ObjectIDFromHex(id)
	}

	data := entities.Promotion{
		Id:          primitive.NewObjectID(),
		BranchId:    branchId,
		Code:        form.Code,
		Name:        form.Name,
		Description: form.Description,
		Type:        form.Type,
		Value:       form.Value,
		MinPurchase: form.MinPurchase,
		MaxDiscount: form.MaxDiscount,
		ProductIds:  productIds,
		StartDate:   form.StartDate,
		EndDate:     form.EndDate,
		Status:      constant.ACTIVE,
		CreatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedBy:   form.CreatedBy,
		UpdatedDate: time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *promotionEntity) GetPromotions(branchId string) ([]entities.Promotion, error) {
	logrus.Info("GetPromotions")
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
	var results []entities.Promotion
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.Promotion{}
	}
	return results, nil
}

func (entity *promotionEntity) GetPromotionById(id string) (*entities.Promotion, error) {
	logrus.Info("GetPromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Promotion{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *promotionEntity) GetPromotionByCode(code string, branchId string) (*entities.Promotion, error) {
	logrus.Info("GetPromotionByCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{
		"code":   code,
		"status": constant.ACTIVE,
	}
	if branchId != "" {
		objId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = objId
	}

	now := time.Now()
	filter["startDate"] = bson.M{"$lte": now}
	filter["endDate"] = bson.M{"$gte": now}

	data := entities.Promotion{}
	err := entity.repo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *promotionEntity) UpdatePromotionById(id string, form request.UpdatePromotion) (*entities.Promotion, error) {
	logrus.Info("UpdatePromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	productIds := make([]primitive.ObjectID, len(form.ProductIds))
	for i, pid := range form.ProductIds {
		productIds[i], _ = primitive.ObjectIDFromHex(pid)
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"name":        form.Name,
		"description": form.Description,
		"type":        form.Type,
		"value":       form.Value,
		"minPurchase": form.MinPurchase,
		"maxDiscount": form.MaxDiscount,
		"productIds":  productIds,
		"startDate":   form.StartDate,
		"endDate":     form.EndDate,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.Promotion{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *promotionEntity) RemovePromotionById(id string) (*entities.Promotion, error) {
	logrus.Info("RemovePromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Promotion{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
