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

type RefillReminder struct {
	PatientId       primitive.ObjectID `bson:"_id" json:"patientId"`
	LastDispensed   time.Time          `bson:"lastDispensed" json:"lastDispensed"`
	EstimatedRefill time.Time          `bson:"estimatedRefill" json:"estimatedRefill"`
}

type IDispensingLog interface {
	CreateDispensingLog(form request.DispensingLog) (*entities.DispensingLog, error)
	GetDispensingLogs(branchId string) ([]entities.DispensingLog, error)
	GetDispensingLogById(id string, branchId string) (*entities.DispensingLog, error)
	GetDispensingLogsByPatientId(patientId string, branchId string) ([]entities.DispensingLog, error)
	GetDispensingLogsByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.DispensingLog, error)
	GetRefillReminders(branchId string, refillDays int) ([]RefillReminder, error)
}

func NewDispensingLogEntity(resource *db.Resource) IDispensingLog {
	repo := resource.PosDb.Collection("dispensing_logs")
	entity := &dispensingLogEntity{repo: repo}
	ensureDispensingLogIndexes(repo)
	return entity
}

func ensureDispensingLogIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "dispensing_logs branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(repo, "dispensing_logs patientId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "patientId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
}

func (entity *dispensingLogEntity) CreateDispensingLog(form request.DispensingLog) (*entities.DispensingLog, error) {
	logrus.Info("CreateDispensingLog")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
	orderId, err := primitive.ObjectIDFromHex(form.OrderId)
	if err != nil {
		return nil, err
	}
	patientId, err := primitive.ObjectIDFromHex(form.PatientId)
	if err != nil {
		return nil, err
	}

	items := make([]entities.DispensingItem, len(form.Items))
	for i, item := range form.Items {
		productId, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			return nil, err
		}
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
	_, err = entity.repo.InsertOne(ctx, data)
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
	var results []entities.DispensingLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DispensingLog{}
	}
	return results, nil
}

func (entity *dispensingLogEntity) GetDispensingLogById(id string, branchId string) (*entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectId}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
	}
	data := entities.DispensingLog{}
	err = entity.repo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *dispensingLogEntity) GetDispensingLogsByPatientId(patientId string, branchId string) ([]entities.DispensingLog, error) {
	logrus.Info("GetDispensingLogsByPatientId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(patientId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"patientId": objId}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branchObjID
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

func (entity *dispensingLogEntity) GetRefillReminders(branchId string, refillDays int) ([]RefillReminder, error) {
	logrus.Info("GetRefillReminders")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if refillDays <= 0 {
		refillDays = 30
	}

	matchFilter := bson.M{
		"patientId": bson.M{"$ne": primitive.NilObjectID},
	}
	if branchId != "" {
		objId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = objId
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id":           "$patientId",
			"lastDispensed": bson.M{"$max": "$createdDate"},
		}},
		{"$addFields": bson.M{
			"estimatedRefill": bson.M{
				"$dateAdd": bson.M{
					"startDate": "$lastDispensed",
					"unit":      "day",
					"amount":    refillDays,
				},
			},
		}},
		{"$match": bson.M{
			"estimatedRefill": bson.M{"$lte": time.Now().AddDate(0, 0, 7)},
		}},
		{"$sort": bson.M{"estimatedRefill": 1}},
	}

	var results []RefillReminder
	cursor, err := entity.repo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []RefillReminder{}
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
	var results []entities.DispensingLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DispensingLog{}
	}
	return results, nil
}
