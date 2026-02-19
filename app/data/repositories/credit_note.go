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

type creditNoteEntity struct {
	repo *mongo.Collection
}

type ICreditNote interface {
	CreateCreditNote(form request.CreditNote) (*entities.CreditNote, error)
	GetCreditNotes(branchId string) ([]entities.CreditNote, error)
	GetCreditNoteById(id string) (*entities.CreditNote, error)
	UpdateCreditNoteById(id string, form request.UpdateCreditNote) (*entities.CreditNote, error)
	RemoveCreditNoteById(id string) (*entities.CreditNote, error)
}

func NewCreditNoteEntity(resource *db.Resource) ICreditNote {
	repo := resource.PosDb.Collection("credit_notes")
	entity := &creditNoteEntity{repo: repo}
	ensureCreditNoteIndexes(repo)
	return entity
}

func ensureCreditNoteIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create credit_notes index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create credit_notes code index: ", err)
	}
}

func (entity *creditNoteEntity) CreateCreditNote(form request.CreditNote) (*entities.CreditNote, error) {
	logrus.Info("CreateCreditNote")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	orderId, _ := primitive.ObjectIDFromHex(form.OrderId)

	items := make([]entities.CreditNoteItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.CreditNoteItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
			StockId:   item.StockId,
		}
	}

	data := entities.CreditNote{
		Id:          primitive.NewObjectID(),
		BranchId:    branchId,
		OrderId:     orderId,
		Code:        form.Code,
		Reason:      form.Reason,
		Items:       items,
		TotalRefund: form.TotalRefund,
		Note:        form.Note,
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

func (entity *creditNoteEntity) GetCreditNotes(branchId string) ([]entities.CreditNote, error) {
	logrus.Info("GetCreditNotes")
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
	var results []entities.CreditNote
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.CreditNote{}
	}
	return results, nil
}

func (entity *creditNoteEntity) GetCreditNoteById(id string) (*entities.CreditNote, error) {
	logrus.Info("GetCreditNoteById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.CreditNote{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *creditNoteEntity) UpdateCreditNoteById(id string, form request.UpdateCreditNote) (*entities.CreditNote, error) {
	logrus.Info("UpdateCreditNoteById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	items := make([]entities.CreditNoteItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.CreditNoteItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
			StockId:   item.StockId,
		}
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"reason":      form.Reason,
		"items":       items,
		"totalRefund": form.TotalRefund,
		"note":        form.Note,
		"updatedBy":   form.UpdatedBy,
		"updatedDate": time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.CreditNote{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *creditNoteEntity) RemoveCreditNoteById(id string) (*entities.CreditNote, error) {
	logrus.Info("RemoveCreditNoteById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.CreditNote{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
