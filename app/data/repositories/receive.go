package repositories

import (
	"context"
	"errors"
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
	client           *mongo.Client
	receiveRepo      *mongo.Collection
	receiveItemsRepo *mongo.Collection
}

type IReceive interface {
	GetReceives(form request.GetReceiveRange) ([]entities.Receive, error)
	CreateReceive(form request.Receive) (*entities.Receive, error)
	GetReceiveById(id string, branchId string) (*entities.Receive, error)
	RemoveReceiveById(id string) (*entities.Receive, error)
	UpdateReceiveById(id string, form request.UpdateReceive) (*entities.Receive, error)
	UpdateReceiveTotalCostById(id string, totalCost float64) (*entities.Receive, error)
	UpdateReceiveItemsById(id string, form request.UpdateReceiveItems) (*entities.Receive, error)
	CreateReceiveItem(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error)
	GetReceiveItemsByReceiveId(receiveId string) ([]entities.ReceiveItem, error)
	GetReceiveItemsByReceiveIds(receiveIds []string) ([]entities.ReceiveItem, error)
	GetReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error)
	RemoveReceiveItemByLotId(lotId string) (*entities.ReceiveItem, error)
	DeleteReceiveItemsByReceiveId(receiveId string) error
	CancelReceiveById(id string, updatedBy string) (*entities.Receive, error)
}

// ErrReceiveLocked is a change to a Receive that is no longer a draft: once
// imported its Stock exists, and once cancelled it is closed.
var ErrReceiveLocked = errors.New("receive is imported or cancelled and can no longer change")

// draft matches the Receive only while it can still change. The Stock ledger
// imports under the same condition, so an edit or cancel and an import cannot
// both win.
func draft(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id, "status": bson.M{"$nin": bson.A{constant.IMPORTED, constant.CANCELLED}}}
}

// changeDraft applies update to the Receive while it is a draft. It tells a
// Receive that does not exist (mongo.ErrNoDocuments) from one that is locked.
func (entity *receiveEntity) changeDraft(ctx context.Context, id primitive.ObjectID, update bson.M) (*entities.Receive, error) {
	after := options.After
	data := entities.Receive{}
	err := entity.receiveRepo.FindOneAndUpdate(ctx, draft(id), update,
		&options.FindOneAndUpdateOptions{ReturnDocument: &after}).Decode(&data)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if n, cErr := entity.receiveRepo.CountDocuments(ctx, bson.M{"_id": id}); cErr == nil && n > 0 {
			return nil, ErrReceiveLocked
		}
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// replaceItems swaps a draft Receive's items. Called after changeDraft in the
// same transaction, so a locked Receive's items are never touched.
func (entity *receiveEntity) replaceItems(ctx context.Context, id primitive.ObjectID, items []entities.ReceiveItem) error {
	if _, err := entity.receiveItemsRepo.DeleteMany(ctx, bson.M{"receiveId": id}); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(items))
	for _, item := range items {
		item.ReceiveId = id
		docs = append(docs, item)
	}
	_, err := entity.receiveItemsRepo.InsertMany(ctx, docs)
	return err
}

func NewReceiveEntity(resource *db.Resource) IReceive {
	receiveRepo := resource.PosDb.Collection("receives")
	receiveItemsRepo := resource.PosDb.Collection("receive_items")
	entity := &receiveEntity{
		client:           resource.Client,
		receiveRepo:      receiveRepo,
		receiveItemsRepo: receiveItemsRepo,
	}
	ensureReceiveIndexes(receiveRepo, receiveItemsRepo)
	return entity
}

func ensureReceiveIndexes(receiveRepo *mongo.Collection, receiveItemsRepo *mongo.Collection) {
	createCollectionIndex(receiveRepo, "receives createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(receiveItemsRepo, "receive_items receiveId", mongo.IndexModel{
		Keys: bson.D{{Key: "receiveId", Value: 1}},
	})
	createCollectionIndex(receiveItemsRepo, "receive_items lotId", mongo.IndexModel{
		Keys: bson.D{{Key: "lotId", Value: 1}},
	})
	createCollectionIndex(receiveRepo, "receives branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *receiveEntity) GetReceives(form request.GetReceiveRange) (items []entities.Receive, err error) {
	logrus.Info("GetReceives")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{
		"createdDate": bson.M{
			"$gt": form.StartDate.Time,
			"$lt": form.EndDate.Time,
		},
	}
	if form.BranchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(form.BranchId)
		if err != nil {
			return nil, err
		}
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
	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
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

func (entity *receiveEntity) GetReceiveById(id string, branchId string) (*entities.Receive, error) {
	logrus.Info("GetReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter, err := inBranch(id, branchId)
	if err != nil {
		return nil, err
	}
	data := entities.Receive{}
	err = entity.receiveRepo.FindOne(ctx, filter).Decode(&data)
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

	return entity.changeDraft(ctx, obId, bson.M{"$set": bson.M{
		"totalCost":   totalCost,
		"updatedDate": time.Now(),
	}})
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
	items, err := buildReceiveItems(form.ReceiveItems)
	if err != nil {
		return nil, err
	}
	data, err := entity.changeDraft(ctx, obId, bson.M{"$set": bson.M{
		"supplierId":  supplier,
		"reference":   form.Reference,
		"totalCost":   calculateReceiveItemsTotalCost(items),
		"items":       items,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}})
	if err != nil {
		return nil, err
	}
	if err := entity.replaceItems(ctx, obId, items); err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *receiveEntity) UpdateReceiveItemsById(id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
	logrus.Info("UpdateReceiveItems")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.updateReceiveItemsByIdWithContext(ctx, id, form)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Receive
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.updateReceiveItemsByIdWithContext(sessCtx, id, form)
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

func (entity *receiveEntity) updateReceiveItemsByIdWithContext(ctx context.Context, id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	items, err := buildReceiveItems(form.ReceiveItems)
	if err != nil {
		return nil, err
	}
	data, err := entity.changeDraft(ctx, obId, bson.M{"$set": bson.M{
		"items":       items,
		"totalCost":   calculateReceiveItemsTotalCost(items),
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}})
	if err != nil {
		return nil, err
	}
	if err := entity.replaceItems(ctx, obId, items); err != nil {
		return nil, err
	}
	return data, nil
}

func buildReceiveItems(items []request.ReceiveItem) ([]entities.ReceiveItem, error) {
	result := make([]entities.ReceiveItem, 0, len(items))
	for _, item := range items {
		productId, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			return nil, err
		}
		receiveItem := entities.ReceiveItem{
			ProductId:    productId,
			CostPrice:    item.CostPrice,
			Quantity:     item.Quantity,
			LotNumber:    item.LotNumber,
			UnitId:       item.UnitId,
			BaseQuantity: item.BaseQuantity,
		}
		if item.ExpireDate != "" {
			if t, e := time.Parse(time.RFC3339, item.ExpireDate); e == nil {
				receiveItem.ExpireDate = t
			}
		}
		result = append(result, receiveItem)
	}
	return result, nil
}

func calculateReceiveItemsTotalCost(items []entities.ReceiveItem) float64 {
	var totalCost float64
	for _, item := range items {
		totalCost += item.CostPrice * float64(item.Quantity)
	}
	return totalCost
}

func (entity *receiveEntity) CreateReceiveItem(receiveId string, _ string, productId string, form request.Product) (*entities.ReceiveItem, error) {
	logrus.Info("CreateReceiveItem")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	recvId, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return nil, err
	}
	data := entities.ReceiveItem{
		ReceiveId:  recvId,
		ProductId:  product,
		Quantity:   form.Quantity,
		CostPrice:  form.CostPrice,
		LotNumber:  form.LotNumber,
		ExpireDate: form.ExpireDate.Time,
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

func (entity *receiveEntity) GetReceiveItemsByReceiveIds(receiveIds []string) (items []entities.ReceiveItem, err error) {
	logrus.Info("GetReceiveItemsByReceiveIds")
	if len(receiveIds) == 0 {
		return []entities.ReceiveItem{}, nil
	}
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectIDs := make([]primitive.ObjectID, 0, len(receiveIds))
	for _, id := range receiveIds {
		receiveID, convErr := primitive.ObjectIDFromHex(id)
		if convErr != nil {
			return nil, convErr
		}
		objectIDs = append(objectIDs, receiveID)
	}

	cursor, err := entity.receiveItemsRepo.Find(ctx, bson.M{"receiveId": bson.M{"$in": objectIDs}})
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

// CancelReceiveById closes a draft Receive; an imported one stays imported.
func (entity *receiveEntity) CancelReceiveById(id string, updatedBy string) (*entities.Receive, error) {
	logrus.Info("CancelReceiveById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	obId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return entity.changeDraft(ctx, obId, bson.M{"$set": bson.M{
		"status":      constant.CANCELLED,
		"updatedBy":   updatedBy,
		"updatedDate": time.Now(),
	}})
}
