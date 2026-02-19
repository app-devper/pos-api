package repositories

import (
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

type dispensingLogEntity struct {
	repo *mongo.Collection
}

type IDispensingLog interface {
	CreateDispensingLog(form request.DispensingLog) (*entities.DispensingLog, error)
	GetDispensingLogs(branchId string) ([]entities.DispensingLog, error)
	GetDispensingLogById(id string) (*entities.DispensingLog, error)
	GetDispensingLogsByPatientId(patientId string) ([]entities.DispensingLog, error)
	GetDispensingLogsByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.DispensingLog, error)
}

func NewDispensingLogEntity(resource *db.Resource) IDispensingLog {
	repo := resource.PosDb.Collection("dispensing_logs")
	entity := &dispensingLogEntity{repo: repo}
	ensureDispensingLogIndexes(repo)
	return entity
}

func ensureDispensingLogIndexes(repo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	_, err := repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create dispensing_logs index: ", err)
	}
	_, err = repo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "patientId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create dispensing_logs patientId index: ", err)
	}
}

func (entity *dispensingLogEntity) CreateDispensingLog(form request.DispensingLog) (*entities.DispensingLog, error) {
	logrus.Info("CreateDispensingLog")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, _ := primitive.ObjectIDFromHex(form.BranchId)
	orderId, _ := primitive.ObjectIDFromHex(form.OrderId)
	patientId, _ := primitive.ObjectIDFromHex(form.PatientId)

	items := make([]entities.DispensingItem, len(form.Items))
	for i, item := range form.Items {
		productId, _ := primitive.ObjectIDFromHex(item.ProductId)
		items[i] = entities.DispensingItem{
			ProductId:   productId,
			ProductName: item.ProductName,
			GenericName: item.GenericName,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			Dosage:      item.Dosage,
			LotNumber:   item.LotNumber,
		}
	}

	data := entities.DispensingLog{
		Id:             primitive.NewObjectID(),
		BranchId:       branchId,
		OrderId:        orderId,
		PatientId:      patientId,
		Items:          items,
		PharmacistName: form.PharmacistName,
		LicenseNo:      form.LicenseNo,
		Note:           form.Note,
		CreatedBy:      form.CreatedBy,
		CreatedDate:    time.Now(),
	}
	_, err := entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *dispensingLogEntity) GetDispensingLogs(branchId string) ([]entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogs")
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
	var results []entities.DispensingLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DispensingLog{}
	}
	return results, nil
}

func (entity *dispensingLogEntity) GetDispensingLogById(id string) (*entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.DispensingLog{}
	err = entity.repo.FindOne(ctx, bson.M{"_id": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *dispensingLogEntity) GetDispensingLogsByPatientId(patientId string) ([]entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogsByPatientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(patientId)
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, bson.M{"patientId": objId}, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.DispensingLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DispensingLog{}
	}
	return results, nil
}

func (entity *dispensingLogEntity) GetDispensingLogsByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogsByDateRange")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{
		"createdDate": bson.M{"$gte": startDate, "$lte": endDate},
	}
	if branchId != "" {
		objId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = objId
	}
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.DispensingLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DispensingLog{}
	}
	return results, nil
}
