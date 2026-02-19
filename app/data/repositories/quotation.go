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

type quotationEntity struct {
	repo *mongo.Collection
}

type IQuotation interface {
	CreateQuotation(form request.Quotation) (*entities.Quotation, error)
	GetQuotations(branchId string) ([]entities.Quotation, error)
	GetQuotationById(id string) (*entities.Quotation, error)
	UpdateQuotationById(id string, form request.UpdateQuotation) (*entities.Quotation, error)
	RemoveQuotationById(id string) (*entities.Quotation, error)
}

func NewQuotationEntity(resource *db.Resource) IQuotation {
	repo := resource.PosDb.Collection("quotations")
	entity := &quotationEntity{repo: repo}
	ensureQuotationIndexes(repo)
	return entity
}

func ensureQuotationIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create quotations index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create quotations code index: ", err)
	}
}

func (entity *quotationEntity) CreateQuotation(form request.Quotation) (*entities.Quotation, error) {
	logrus.Info("CreateQuotation")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)

	items := make([]entities.QuotationItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.QuotationItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Total:     item.Total,
		}
	}

	data := entities.Quotation{
		Id:           primitive.NewObjectID(),
		BranchId:     branchId,
		CustomerCode: form.CustomerCode,
		CustomerName: form.CustomerName,
		Code:         form.Code,
		Items:        items,
		TotalAmount:  form.TotalAmount,
		Discount:     form.Discount,
		Note:         form.Note,
		Status:       constant.ACTIVE,
		ValidUntil:   form.ValidUntil,
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

func (entity *quotationEntity) GetQuotations(branchId string) ([]entities.Quotation, error) {
	logrus.Info("GetQuotations")
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
	var results []entities.Quotation
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.Quotation{}
	}
	return results, nil
}

func (entity *quotationEntity) GetQuotationById(id string) (*entities.Quotation, error) {
	logrus.Info("GetQuotationById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Quotation{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *quotationEntity) UpdateQuotationById(id string, form request.UpdateQuotation) (*entities.Quotation, error) {
	logrus.Info("UpdateQuotationById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	items := make([]entities.QuotationItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.QuotationItem{
			ProductId: productId,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Total:     item.Total,
		}
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	update := bson.M{
		"customerCode": form.CustomerCode,
		"customerName": form.CustomerName,
		"items":        items,
		"totalAmount":  form.TotalAmount,
		"discount":     form.Discount,
		"note":         form.Note,
		"validUntil":   form.ValidUntil,
		"updatedBy":    form.UpdatedBy,
		"updatedDate":  time.Now(),
	}
	if form.Status != "" {
		update["status"] = form.Status
	}

	data := entities.Quotation{}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objectId}, bson.M{"$set": update}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *quotationEntity) RemoveQuotationById(id string) (*entities.Quotation, error) {
	logrus.Info("RemoveQuotationById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Quotation{}
	err = entity.repo.FindOneAndDelete(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
