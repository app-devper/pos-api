package repositories

import (
	"context"
	"fmt"
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

type receiveEntity struct {
	client             *mongo.Client
	receiveRepo        *mongo.Collection
	receiveItemsRepo   *mongo.Collection
	productsRepo       *mongo.Collection
	productUnitsRepo   *mongo.Collection
	productStockRepo   *mongo.Collection
	productHistoryRepo *mongo.Collection
}

type IReceive interface {
	GetReceives(form request.GetReceiveRange) ([]entities.Receive, error)
	CreateReceive(form request.Receive) (*entities.Receive, error)
	GetReceiveById(id string) (*entities.Receive, error)
	RemoveReceiveById(id string) (*entities.Receive, error)
	UpdateReceiveById(id string, form request.UpdateReceive) (*entities.Receive, error)
	UpdateReceiveTotalCostById(id string, totalCost float64) (*entities.Receive, error)
	UpdateReceiveItemsById(id string, form request.UpdateReceiveItems) (*entities.Receive, error)
	CreateReceiveItem(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error)
	GetReceiveItemsByReceiveId(receiveId string) ([]entities.ReceiveItem, error)
	GetReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error)
	RemoveReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error)
	DeleteReceiveItemsByReceiveId(receiveId string) error
	UpdateReceiveStatusById(id string, status string, updatedBy string) (*entities.Receive, error)
	ImportReceiveToStock(receiveId string, userId string, branchId string) (*entities.Receive, error)
}

func NewReceiveEntity(resource *db.Resource) IReceive {
	receiveRepo := resource.PosDb.Collection("receives")
	receiveItemsRepo := resource.PosDb.Collection("receive_items")
	productsRepo := resource.PosDb.Collection("products")
	productUnitsRepo := resource.PosDb.Collection("product_units")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	productHistoryRepo := resource.PosDb.Collection("product_histories")
	entity := &receiveEntity{
		client:             resource.Client,
		receiveRepo:        receiveRepo,
		receiveItemsRepo:   receiveItemsRepo,
		productsRepo:       productsRepo,
		productUnitsRepo:   productUnitsRepo,
		productStockRepo:   productStockRepo,
		productHistoryRepo: productHistoryRepo,
	}
	ensureReceiveIndexes(receiveRepo, receiveItemsRepo)
	return entity
}

func ensureReceiveIndexes(receiveRepo *mongo.Collection, receiveItemsRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := receiveRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create createdDate index: ", err)
	}

	_, err = receiveItemsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "receiveId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create receiveId index: ", err)
	}

	_, err = receiveItemsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "lotId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create lotId index: ", err)
	}

	_, err = receiveRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create receives branchId+createdDate index: ", err)
	}
}

func (entity *receiveEntity) GetReceives(form request.GetReceiveRange) (items []entities.Receive, err error) {
	logrus.Info("GetReceives")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{
		"createdDate": bson.M{
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		},
	}
	if form.BranchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(form.BranchId)
		filter["branchId"] = branchObjId
	}
	cursor, err := entity.receiveRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	items = []entities.Receive{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *receiveEntity) CreateReceive(form request.Receive) (*entities.Receive, error) {
	logrus.Info("CreateReceive")
	ctx, cancel := utils.InitContext()
	defer cancel()
	supplier, err := primitive.ObjectIDFromHex(form.SupplierId)
	if err != nil {
		return nil, err
	}
	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	data := entities.Receive{
		Id:          primitive.NewObjectID(),
		BranchId:    branchId,
		Code:        form.Code,
		Reference:   form.Reference,
		SupplierId:  supplier,
		Items:       []entities.ReceiveItem{},
		Status:      constant.ACTIVE,
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

func (entity *receiveEntity) GetReceiveById(id string) (*entities.Receive, error) {
	logrus.Info("GetReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Receive{}
	err = entity.receiveRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) RemoveReceiveById(id string) (*entities.Receive, error) {
	logrus.Info("RemoveReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.removeReceiveByIdWithContext(ctx, id)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Receive
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.removeReceiveByIdWithContext(sessCtx, id)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *receiveEntity) removeReceiveByIdWithContext(ctx context.Context, id string) (*entities.Receive, error) {
	data := entities.Receive{}
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.receiveRepo.FindOneAndDelete(ctx, bson.M{"_id": obId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.receiveItemsRepo.DeleteMany(ctx, bson.M{"receiveId": obId})
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) UpdateReceiveTotalCostById(id string, totalCost float64) (*entities.Receive, error) {
	logrus.Info("UpdateReceiveTotalCostById")
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
	data := entities.Receive{}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": bson.M{
		"totalCost":   totalCost,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) UpdateReceiveById(id string, form request.UpdateReceive) (*entities.Receive, error) {
	logrus.Info("UpdateReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.updateReceiveByIdWithContext(ctx, id, form)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Receive
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.updateReceiveByIdWithContext(sessCtx, id, form)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *receiveEntity) updateReceiveByIdWithContext(ctx context.Context, id string, form request.UpdateReceive) (*entities.Receive, error) {
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	supplier, err := primitive.ObjectIDFromHex(form.SupplierId)
	if err != nil {
		return nil, err
	}
	items := make([]entities.ReceiveItem, 0, len(form.ReceiveItems))
	for _, item := range form.ReceiveItems {
		productId, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			return nil, err
		}
		ri := entities.ReceiveItem{
			ProductId:    productId,
			CostPrice:    item.CostPrice,
			Quantity:     item.Quantity,
			LotNumber:    item.LotNumber,
			UnitId:       item.UnitId,
			BaseQuantity: item.BaseQuantity,
		}
		if item.ExpireDate != "" {
			if t, e := time.Parse(time.RFC3339, item.ExpireDate); e == nil {
				ri.ExpireDate = t
			}
		}
		items = append(items, ri)
	}

	if _, err = entity.receiveItemsRepo.DeleteMany(ctx, bson.M{"receiveId": obId}); err != nil {
		return nil, err
	}
	if len(items) > 0 {
		docs := make([]interface{}, 0, len(items))
		for _, item := range items {
			item.ReceiveId = obId
			docs = append(docs, item)
		}
		if _, err = entity.receiveItemsRepo.InsertMany(ctx, docs); err != nil {
			return nil, err
		}
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	data := entities.Receive{}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": bson.M{
		"supplierId":  supplier,
		"reference":   form.Reference,
		"totalCost":   form.TotalCost,
		"items":       items,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) UpdateReceiveItemsById(id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
	logrus.Info("UpdateReceiveItems")
	ctx, cancel := utils.InitContext()
	defer cancel()
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	items := make([]entities.ReceiveItem, 0, len(form.ReceiveItems))
	for _, item := range form.ReceiveItems {
		productId, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			return nil, err
		}
		ri := entities.ReceiveItem{
			ProductId:    productId,
			CostPrice:    item.CostPrice,
			Quantity:     item.Quantity,
			LotNumber:    item.LotNumber,
			UnitId:       item.UnitId,
			BaseQuantity: item.BaseQuantity,
		}
		if item.ExpireDate != "" {
			if t, e := time.Parse(time.RFC3339, item.ExpireDate); e == nil {
				ri.ExpireDate = t
			}
		}
		items = append(items, ri)
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	data := entities.Receive{}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": bson.M{
		"items":       items,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) CreateReceiveItem(receiveId string, _ string, productId string, form request.Product) (*entities.ReceiveItem, error) {
	logrus.Info("CreateReceiveItem")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	recvId, _ := primitive.ObjectIDFromHex(receiveId)
	data := entities.ReceiveItem{
		ReceiveId:  recvId,
		ProductId:  product,
		Quantity:   form.Quantity,
		CostPrice:  form.CostPrice,
		LotNumber:  form.LotNumber,
		ExpireDate: form.ExpireDate,
	}
	_, err = entity.receiveItemsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) GetReceiveItemsByReceiveId(receiveId string) (items []entities.ReceiveItem, err error) {
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
	items = []entities.ReceiveItem{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *receiveEntity) GetReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error) {
	logrus.Info("GetReceiveItemByLotId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	lot, err := primitive.ObjectIDFromHex(lotId)
	if err != nil {
		return nil, err
	}
	data := entities.ReceiveItem{}
	err = entity.receiveItemsRepo.FindOne(ctx, bson.M{"lotId": lot}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) RemoveReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error) {
	logrus.Info("RemoveReceiveItemByLotId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ReceiveItem{}
	lot, err := primitive.ObjectIDFromHex(lotId)
	if err != nil {
		return nil, err
	}
	err = entity.receiveItemsRepo.FindOneAndDelete(ctx, bson.M{"lotId": lot}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) DeleteReceiveItemsByReceiveId(receiveId string) error {
	logrus.Info("DeleteReceiveItemsByReceiveId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	recvId, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return err
	}
	_, err = entity.receiveItemsRepo.DeleteMany(ctx, bson.M{"receiveId": recvId})
	return err
}

func (entity *receiveEntity) UpdateReceiveStatusById(id string, status string, updatedBy string) (*entities.Receive, error) {
	logrus.Info("UpdateReceiveStatusById")
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
	data := entities.Receive{}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": obId}, bson.M{"$set": bson.M{
		"status":      status,
		"updatedBy":   updatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *receiveEntity) ImportReceiveToStock(receiveId string, userId string, branchId string) (*entities.Receive, error) {
	logrus.Info("ImportReceiveToStock")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.importReceiveToStockWithContext(ctx, receiveId, userId, branchId)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Receive
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.importReceiveToStockWithContext(sessCtx, receiveId, userId, branchId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *receiveEntity) importReceiveToStockWithContext(ctx context.Context, receiveId string, userId string, branchId string) (*entities.Receive, error) {
	receiveObjID, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return nil, err
	}
	branchObjID, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}

	receive := entities.Receive{}
	if err := entity.receiveRepo.FindOne(ctx, bson.M{"_id": receiveObjID}).Decode(&receive); err != nil {
		return nil, err
	}
	if receive.Status == constant.IMPORTED {
		return nil, fmt.Errorf("receive already imported")
	}

	cursor, err := entity.receiveItemsRepo.Find(ctx, bson.M{"receiveId": receiveObjID})
	if err != nil {
		return nil, err
	}
	items := []entities.ReceiveItem{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	var totalCost float64
	now := time.Now()
	for _, item := range items {
		product := entities.Product{}
		if err := entity.productsRepo.FindOne(ctx, bson.M{"_id": item.ProductId}).Decode(&product); err != nil {
			return nil, fmt.Errorf("failed to import receive item for product %s: %w", item.ProductId.Hex(), err)
		}

		unit := entities.ProductUnit{}
		if err := entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": item.ProductId, "unit": product.Unit}).Decode(&unit); err != nil {
			return nil, fmt.Errorf("failed to import receive item for product %s: %w", item.ProductId.Hex(), err)
		}

		if item.Quantity > 0 {
			sequence, err := entity.getNextReceiveProductStockSequence(ctx, item.ProductId, unit.Id)
			if err != nil {
				return nil, err
			}

			stock := entities.ProductStock{
				Id:          primitive.NewObjectID(),
				BranchId:    branchObjID,
				ProductId:   item.ProductId,
				UnitId:      unit.Id,
				ReceiveCode: receive.Code,
				Sequence:    sequence,
				LotNumber:   item.LotNumber,
				CostPrice:   item.CostPrice,
				Import:      item.Quantity,
				Quantity:    item.Quantity,
				ExpireDate:  item.ExpireDate,
				ImportDate:  now,
			}
			if _, err := entity.productStockRepo.InsertOne(ctx, stock); err != nil {
				return nil, fmt.Errorf("failed to create stock for product %s: %w", item.ProductId.Hex(), err)
			}

			balance, err := entity.getReceiveProductStockBalance(ctx, item.ProductId, unit.Id, branchObjID)
			if err != nil {
				return nil, err
			}

			history := request.AddProductStockHistory(item.ProductId.Hex(), product.Unit, request.ProductStock{
				ProductId:   item.ProductId.Hex(),
				UnitId:      unit.Id.Hex(),
				ReceiveCode: receive.Code,
				Quantity:    item.Quantity,
				CostPrice:   item.CostPrice,
				LotNumber:   item.LotNumber,
				ExpireDate:  item.ExpireDate,
				ImportDate:  now,
				UpdatedBy:   userId,
				BranchId:    branchId,
			}, balance)
			history.BranchId = branchId

			historyDoc := entities.ProductHistory{
				Id:          primitive.NewObjectID(),
				BranchId:    branchObjID,
				ProductId:   item.ProductId,
				Type:        history.Type,
				Description: history.Description,
				Unit:        history.Unit,
				Import:      history.Import,
				Quantity:    history.Quantity,
				CostPrice:   history.CostPrice,
				Price:       history.Price,
				Balance:     history.Balance,
				CreatedBy:   history.CreatedBy,
				CreatedDate: now,
			}
			if _, err := entity.productHistoryRepo.InsertOne(ctx, historyDoc); err != nil {
				return nil, fmt.Errorf("failed to create stock history for product %s: %w", item.ProductId.Hex(), err)
			}
		}

		totalCost += item.CostPrice * float64(item.Quantity)
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}
	result := entities.Receive{}
	err = entity.receiveRepo.FindOneAndUpdate(ctx, bson.M{"_id": receiveObjID}, bson.M{"$set": bson.M{
		"totalCost":   totalCost,
		"status":      constant.IMPORTED,
		"updatedBy":   userId,
		"updatedDate": now,
	}}, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	result.Items = items
	return &result, nil
}

func (entity *receiveEntity) getNextReceiveProductStockSequence(ctx context.Context, productId primitive.ObjectID, unitId primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"productId": productId, "unitId": unitId}},
		{"$group": bson.M{"_id": nil, "maxSequence": bson.M{"$max": "$sequence"}}},
	}
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	switch v := results[0]["maxSequence"].(type) {
	case int32:
		return int(v) + 1, nil
	case int64:
		return int(v) + 1, nil
	case float64:
		return int(v) + 1, nil
	default:
		return 1, nil
	}
}

func (entity *receiveEntity) getReceiveProductStockBalance(ctx context.Context, productId primitive.ObjectID, unitId primitive.ObjectID, branchId primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"productId": productId, "unitId": unitId, "branchId": branchId}},
		{"$group": bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}},
	}
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	switch v := results[0]["balance"].(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, nil
	}
}
