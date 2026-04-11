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
	GetPromotionById(id string, branchId string) (*entities.Promotion, error)
	GetPromotionByCode(code string, branchId string) (*entities.Promotion, error)
	UpdatePromotionById(id string, branchId string, form request.UpdatePromotion) (*entities.Promotion, error)
	RemovePromotionById(id string, branchId string, updatedBy string) (*entities.Promotion, error)
}

func NewPromotionEntity(resource *db.Resource) IPromotion {
	repo := resource.PosDb.Collection("promotions")
	entity := &promotionEntity{repo: repo}
	ensurePromotionIndexes(repo)
	return entity
}

func ensurePromotionIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "promotions branchId+status", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "status", Value: 1}},
	})
	createCollectionIndex(repo, "promotions code+branchId", mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}, {Key: "branchId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func (entity *promotionEntity) CreatePromotion(form request.Promotion) (*entities.Promotion, error) {
	logrus.Info("CreatePromotion")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
	productIds := make([]primitive.ObjectID, len(form.ProductIds))
	for i, id := range form.ProductIds {
		productIds[i], err = primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
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
		StartDate:   form.StartDate.Time,
		EndDate:     form.EndDate.Time,
		Status:      constant.ACTIVE,
		CreatedBy:   form.CreatedBy,
		CreatedDate: time.Now(),
		UpdatedBy:   form.CreatedBy,
		UpdatedDate: time.Now(),
	}
	_, err = entity.repo.InsertOne(ctx, data)
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
		objId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
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

func (entity *promotionEntity) GetPromotionById(id string, branchId string) (*entities.Promotion, error) {
	logrus.Info("GetPromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectId}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjId
	}
	data := entities.Promotion{}
	err = entity.repo.FindOne(ctx, filter).Decode(&data)
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
		objId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
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

func (entity *promotionEntity) UpdatePromotionById(id string, branchId string, form request.UpdatePromotion) (*entities.Promotion, error) {
	logrus.Info("UpdatePromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectId}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjId
	}

	productIds := make([]primitive.ObjectID, len(form.ProductIds))
	for i, pid := range form.ProductIds {
		productIds[i], err = primitive.ObjectIDFromHex(pid)
		if err != nil {
			return nil, err
		}
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
		"startDate":   form.StartDate.Time,
		"endDate":     form.EndDate.Time,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.Promotion{}
	err = entity.repo.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *promotionEntity) RemovePromotionById(id string, branchId string, updatedBy string) (*entities.Promotion, error) {
	logrus.Info("RemovePromotionById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectId}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjId
	}
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}
	data := entities.Promotion{}
	err = entity.repo.FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{
		"status":      constant.INACTIVE,
		"updatedBy":   updatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
