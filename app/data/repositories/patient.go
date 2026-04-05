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

type patientEntity struct {
	repo *mongo.Collection
}

type IPatient interface {
	CreatePatient(form request.Patient) (*entities.Patient, error)
	GetPatients(branchId string) ([]entities.Patient, error)
	GetPatientById(id string, branchId string) (*entities.Patient, error)
	GetPatientByCustomerCode(customerCode string, branchId string) (*entities.Patient, error)
	UpdatePatientById(id string, form request.UpdatePatient, branchId string) (*entities.Patient, error)
	RemovePatientById(id string, branchId string) (*entities.Patient, error)
}

func NewPatientEntity(resource *db.Resource) IPatient {
	repo := resource.PosDb.Collection("patients")
	entity := &patientEntity{repo: repo}
	ensurePatientIndexes(repo)
	return entity
}

func ensurePatientIndexes(repo *mongo.Collection) {
	createCollectionIndex(repo, "patients customerCode+branchId", mongo.IndexModel{
		Keys:    bson.D{{Key: "customerCode", Value: 1}, {Key: "branchId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func toEntityAllergies(items []request.PatientDrugAllergy) []entities.DrugAllergy {
	result := make([]entities.DrugAllergy, len(items))
	for i, a := range items {
		result[i] = entities.DrugAllergy{
			DrugName: a.DrugName,
			Reaction: a.Reaction,
			Severity: a.Severity,
		}
	}
	return result
}

func (entity *patientEntity) CreatePatient(form request.Patient) (*entities.Patient, error) {
	logrus.Info("CreatePatient")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}
	data := entities.Patient{
		Id:                 primitive.NewObjectID(),
		BranchId:           branchId,
		CustomerCode:       form.CustomerCode,
		IdCard:             form.IdCard,
		DateOfBirth:        form.DateOfBirth,
		Gender:             form.Gender,
		BloodType:          form.BloodType,
		Weight:             form.Weight,
		Allergies:          toEntityAllergies(form.Allergies),
		ChronicDiseases:    form.ChronicDiseases,
		CurrentMedications: form.CurrentMedications,
		Note:               form.Note,
		CreatedBy:          form.CreatedBy,
		CreatedDate:        time.Now(),
		UpdatedBy:          form.CreatedBy,
		UpdatedDate:        time.Now(),
	}
	if data.Allergies == nil {
		data.Allergies = []entities.DrugAllergy{}
	}
	if data.ChronicDiseases == nil {
		data.ChronicDiseases = []string{}
	}
	if data.CurrentMedications == nil {
		data.CurrentMedications = []string{}
	}
	_, err = entity.repo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *patientEntity) GetPatients(branchId string) ([]entities.Patient, error) {
	logrus.Info("GetPatients")
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
	var results []entities.Patient
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.Patient{}
	}
	return results, nil
}

func (entity *patientEntity) GetPatientById(id string, branchId string) (*entities.Patient, error) {
	logrus.Info("GetPatientById")
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
	data := entities.Patient{}
	err = entity.repo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *patientEntity) GetPatientByCustomerCode(customerCode string, branchId string) (*entities.Patient, error) {
	logrus.Info("GetPatientByCustomerCode")
	ctx, cancel := utils.InitContext()
	defer cancel()

	filter := bson.M{"customerCode": customerCode}
	if branchId != "" {
		objId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = objId
	}
	data := entities.Patient{}
	err := entity.repo.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *patientEntity) UpdatePatientById(id string, form request.UpdatePatient, branchId string) (*entities.Patient, error) {
	logrus.Info("UpdatePatientById")
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

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{ReturnDocument: &isReturnNewDoc}

	data := entities.Patient{}
	err = entity.repo.FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{
		"idCard":             form.IdCard,
		"dateOfBirth":        form.DateOfBirth,
		"gender":             form.Gender,
		"bloodType":          form.BloodType,
		"weight":             form.Weight,
		"allergies":          toEntityAllergies(form.Allergies),
		"chronicDiseases":    form.ChronicDiseases,
		"currentMedications": form.CurrentMedications,
		"note":               form.Note,
		"updatedBy":          form.UpdatedBy,
		"updatedDate":        time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *patientEntity) RemovePatientById(id string, branchId string) (*entities.Patient, error) {
	logrus.Info("RemovePatientById")
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
	data := entities.Patient{}
	err = entity.repo.FindOneAndDelete(ctx, filter).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
