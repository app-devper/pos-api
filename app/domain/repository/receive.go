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

type receiveEntity struct {
	receiveRepo      *mongo.Collection
	receiveItemsRepo *mongo.Collection
}

type IReceive interface {
	GetReceives(form request.GetReceiveRange) ([]model.Receive, error)
	CreateReceive(form request.Receive) (*model.Receive, error)
	GetReceiveById(id string) (*model.Receive, error)
	RemoveReceiveById(id string) (*model.Receive, error)
	UpdateReceiveById(id string, form request.Receive) (*model.Receive, error)
	UpdateReceiveTotalCostById(id string, totalCost float64) (*model.Receive, error)
	CreateReceiveItem(id string, lotId string) (*model.ReceiveItem, error)
	GetReceiveItemsByReceiveId(receiveId string) ([]model.ReceiveItem, error)
	RemoveReceiveItemByLotId(lotId string) (*model.ReceiveItem, error)
}

func NewReceiveEntity(resource *db.Resource) IReceive {
	receiveRepo := resource.PosDb.Collection("receives")
	receiveItemsRepo := resource.PosDb.Collection("receive_items")
	entity := &receiveEntity{
		receiveRepo:      receiveRepo,
		receiveItemsRepo: receiveItemsRepo,
	}
	return entity
}

func (entity *receiveEntity) GetReceives(form request.GetReceiveRange) (items []model.Receive, err error) {
	logrus.Info("GetReceives")
	ctx, cancel := utils.InitContext()
	defer cancel()

	cursor, err := entity.receiveRepo.Find(ctx, bson.M{
		"createdDate": bson.M{
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		},
	})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := model.Receive{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.Receive{}
	}
	return items, nil
}

func (entity *receiveEntity) CreateReceive(form request.Receive) (*model.Receive, error) {
	logrus.Info("CreateReceive")
	ctx, cancel := utils.InitContext()
	defer cancel()
	supplier, err := primitive.ObjectIDFromHex(form.SupplierId)
	if err != nil {
		return nil, err
	}
	data := model.Receive{
		Id:          primitive.NewObjectID(),
		Code:        form.Code,
		Reference:   form.Reference,
		SupplierId:  supplier,
		CreatedBy:   form.UpdatedBy,
		UpdatedBy:   form.UpdatedBy,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}
	_, err = entity.receiveRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) GetReceiveById(id string) (*model.Receive, error) {
	logrus.Info("GetReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := model.Receive{}
	err := entity.receiveRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) RemoveReceiveById(id string) (*model.Receive, error) {
	logrus.Info("RemoveReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.Receive{}
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.receiveRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.receiveRepo.DeleteOne(ctx, bson.M{"_id": obId})
	if err != nil {
		logrus.Error(err)
	}
	_, err = entity.receiveItemsRepo.DeleteMany(ctx, bson.M{"receiveId": obId})
	if err != nil {
		logrus.Error(err)
	}
	return &data, nil
}

func (entity *receiveEntity) UpdateReceiveTotalCostById(id string, totalCost float64) (*model.Receive, error) {
	logrus.Info("UpdateReceiveTotalCostById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := model.Receive{}
	err = entity.receiveRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}

	data.TotalCost = totalCost
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) UpdateReceiveById(id string, form request.Receive) (*model.Receive, error) {
	logrus.Info("UpdateReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := model.Receive{}
	err = entity.receiveRepo.FindOne(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	supplier, err := primitive.ObjectIDFromHex(form.SupplierId)
	if err != nil {
		return nil, err
	}
	data.SupplierId = supplier
	data.Reference = form.Reference
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) CreateReceiveItem(receiveId string, lotId string) (*model.ReceiveItem, error) {
	logrus.Info("CreateReceiveItem")
	ctx, cancel := utils.InitContext()
	defer cancel()
	receive, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return nil, err
	}
	lot, err := primitive.ObjectIDFromHex(lotId)
	if err != nil {
		return nil, err
	}
	data := model.ReceiveItem{
		ReceiveId: receive,
		LotId:     lot,
	}
	_, err = entity.receiveItemsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) GetReceiveItemsByReceiveId(receiveId string) (items []model.ReceiveItem, err error) {
	logrus.Info("GetReceiveItemsByReceiveId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	receive, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.receiveItemsRepo.Find(ctx, bson.M{"receiveId": receive})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := model.ReceiveItem{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ReceiveItem{}
	}
	return items, nil
}

func (entity *receiveEntity) RemoveReceiveItemByLotId(lotId string) (*model.ReceiveItem, error) {
	logrus.Info("RemoveReceiveItemByLotId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ReceiveItem{}
	lot, err := primitive.ObjectIDFromHex(lotId)
	if err != nil {
		return nil, err
	}
	err = entity.receiveItemsRepo.FindOne(ctx, bson.M{"lotId": lot}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.receiveItemsRepo.DeleteOne(ctx, bson.M{"lotId": lot})
	if err != nil {
		logrus.Error(err)
	}
	return &data, nil
}
