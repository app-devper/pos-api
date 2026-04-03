package repositories

import (
	"context"
	"fmt"
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

type stockTransferEntity struct {
	client           *mongo.Client
	repo             *mongo.Collection
	productStockRepo *mongo.Collection
}

type IStockTransfer interface {
	CreateStockTransfer(form request.StockTransfer) (*entities.StockTransfer, error)
	CreateStockTransferWithReservation(form request.StockTransfer) (*entities.StockTransfer, error)
	GetStockTransfers(branchId string) ([]entities.StockTransfer, error)
	GetStockTransferById(id string) (*entities.StockTransfer, error)
	UpdateStockTransferStatus(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error)
	ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error)
	RejectStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error)
}

func NewStockTransferEntity(resource *db.Resource) IStockTransfer {
	repo := resource.PosDb.Collection("stock_transfers")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	entity := &stockTransferEntity{client: resource.Client, repo: repo, productStockRepo: productStockRepo}
	ensureStockTransferIndexes(repo)
	return entity
}

func ensureStockTransferIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "fromBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create stock_transfers fromBranch index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "toBranchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create stock_transfers toBranch index: ", err)
	}
}

func (entity *stockTransferEntity) CreateStockTransfer(form request.StockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("CreateStockTransfer")
	ctx, cancel := utils.InitContext()
	defer cancel()

	return entity.createStockTransferWithContext(ctx, form)
}

func (entity *stockTransferEntity) CreateStockTransferWithReservation(form request.StockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("CreateStockTransferWithReservation")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createStockTransferWithReservationContext(ctx, form)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var created *entities.StockTransfer
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		result, txErr := entity.createStockTransferWithReservationContext(sessCtx, form)
		if txErr != nil {
			return nil, txErr
		}
		created = result
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (entity *stockTransferEntity) createStockTransferWithReservationContext(ctx context.Context, form request.StockTransfer) (*entities.StockTransfer, error) {
	for _, item := range form.Items {
		if item.StockId == "" {
			continue
		}
		stockID, err := primitive.ObjectIDFromHex(item.StockId)
		if err != nil {
			return nil, err
		}
		result, err := entity.productStockRepo.UpdateOne(ctx, bson.M{
			"_id":      stockID,
			"quantity": bson.M{"$gte": item.Quantity},
		}, bson.M{
			"$inc": bson.M{"quantity": -item.Quantity},
		})
		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			return nil, fmt.Errorf("insufficient stock for product %s", item.ProductId)
		}
	}

	return entity.createStockTransferWithContext(ctx, form)
}

func (entity *stockTransferEntity) createStockTransferWithContext(ctx context.Context, form request.StockTransfer) (*entities.StockTransfer, error) {

	fromBranchId, _ := primitive.ObjectIDFromHex(form.FromBranchId)
	toBranchId, _ := primitive.ObjectIDFromHex(form.ToBranchId)

	items := make([]entities.StockTransferItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.StockTransferItem{
			ProductId: productId,
			StockId:   item.StockId,
			Quantity:  item.Quantity,
		}
	}

	data := entities.StockTransfer{
		Id:           primitive.NewObjectID(),
		FromBranchId: fromBranchId,
		ToBranchId:   toBranchId,
		Code:         form.Code,
		Items:        items,
		Note:         form.Note,
		Status:       "PENDING",
		CreatedBy:    form.CreatedBy,
		CreatedDate:  time.Now(),
		UpdatedBy:    form.CreatedBy,
		UpdatedDate:  time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	logrus.Info("ApproveStockTransfer")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.approveStockTransferWithContext(ctx, id, updatedBy)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var updated *entities.StockTransfer
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		result, txErr := entity.approveStockTransferWithContext(sessCtx, id, updatedBy)
		if txErr != nil {
			return nil, txErr
		}
		updated = result
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (entity *stockTransferEntity) approveStockTransferWithContext(ctx context.Context, id string, updatedBy string) (*entities.StockTransfer, error) {
	transfer, err := entity.getPendingStockTransferByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, item := range transfer.Items {
		if item.StockId == "" {
			continue
		}
		sourceStock, err := entity.getProductStockByIDWithContext(ctx, item.StockId)
		if err != nil {
			return nil, err
		}

		sequence, err := entity.getNextProductStockSequence(ctx, item.ProductId, sourceStock.UnitId)
		if err != nil {
			return nil, err
		}

		stock := entities.ProductStock{
			Id:         primitive.NewObjectID(),
			BranchId:   transfer.ToBranchId,
			ProductId:  item.ProductId,
			UnitId:     sourceStock.UnitId,
			Sequence:   sequence,
			LotNumber:  sourceStock.LotNumber,
			CostPrice:  sourceStock.CostPrice,
			Price:      sourceStock.Price,
			Import:     item.Quantity,
			Quantity:   item.Quantity,
			ExpireDate: sourceStock.ExpireDate,
			ImportDate: time.Now(),
		}
		if _, err := entity.productStockRepo.InsertOne(ctx, stock); err != nil {
			return nil, err
		}
	}

	return entity.updateStockTransferStatusWithContext(ctx, id, request.UpdateStockTransfer{
		Status:    "APPROVED",
		UpdatedBy: updatedBy,
	})
}

func (entity *stockTransferEntity) RejectStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	logrus.Info("RejectStockTransfer")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.rejectStockTransferWithContext(ctx, id, updatedBy)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var updated *entities.StockTransfer
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		result, txErr := entity.rejectStockTransferWithContext(sessCtx, id, updatedBy)
		if txErr != nil {
			return nil, txErr
		}
		updated = result
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (entity *stockTransferEntity) rejectStockTransferWithContext(ctx context.Context, id string, updatedBy string) (*entities.StockTransfer, error) {
	transfer, err := entity.getPendingStockTransferByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, item := range transfer.Items {
		if item.StockId == "" {
			continue
		}
		stockID, err := primitive.ObjectIDFromHex(item.StockId)
		if err != nil {
			return nil, err
		}
		if _, err := entity.productStockRepo.UpdateOne(ctx, bson.M{"_id": stockID}, bson.M{
			"$inc": bson.M{"quantity": item.Quantity},
		}); err != nil {
			return nil, err
		}
	}

	return entity.updateStockTransferStatusWithContext(ctx, id, request.UpdateStockTransfer{
		Status:    "REJECTED",
		UpdatedBy: updatedBy,
	})
}

func (entity *stockTransferEntity) GetStockTransfers(branchId string) ([]entities.StockTransfer, error) {
	logrus.Info("GetStockTransfers")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(branchId)
	filter := bson.M{
		"$or": []bson.M{
			{"fromBranchId": objId},
			{"toBranchId": objId},
		},
	}
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.StockTransfer
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.StockTransfer{}
	}
	return results, nil
}

func (entity *stockTransferEntity) GetStockTransferById(id string) (*entities.StockTransfer, error) {
	logrus.Info("GetStockTransferById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.StockTransfer{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) UpdateStockTransferStatus(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error) {
	logrus.Info("UpdateStockTransferStatus")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.updateStockTransferStatusWithContext(ctx, id, form)
}

func (entity *stockTransferEntity) updateStockTransferStatusWithContext(ctx context.Context, id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	data := entities.StockTransfer{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId, "status": "PENDING"}, bson.M{"$set": bson.M{
		"status":      form.Status,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) getPendingStockTransferByIDWithContext(ctx context.Context, id string) (*entities.StockTransfer, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.StockTransfer{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId, "status": "PENDING"}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) getProductStockByIDWithContext(ctx context.Context, id string) (*entities.ProductStock, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.ProductStock{}
	err = entity.productStockRepo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *stockTransferEntity) getNextProductStockSequence(ctx context.Context, productId primitive.ObjectID, unitId primitive.ObjectID) (int, error) {
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
