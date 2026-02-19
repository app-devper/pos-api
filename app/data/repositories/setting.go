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

type settingEntity struct {
	settingRepo *mongo.Collection
}

type ISetting interface {
	GetSettingByBranchId(branchId string) (*entities.Setting, error)
	UpsertSetting(form request.Setting) (*entities.Setting, error)
}

func NewSettingEntity(resource *db.Resource) ISetting {
	settingRepo := resource.PosDb.Collection("settings")
	entity := &settingEntity{settingRepo: settingRepo}
	ensureSettingIndexes(settingRepo)
	return entity
}

func ensureSettingIndexes(settingRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := settingRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "branchId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create settings branchId index: ", err)
	}
}

func (entity *settingEntity) GetSettingByBranchId(branchId string) (*entities.Setting, error) {
	logrus.Info("GetSettingByBranchId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}

	data := entities.Setting{}
	err = entity.settingRepo.FindOne(ctx, bson.M{"branchId": objectId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *settingEntity) UpsertSetting(form request.Setting) (*entities.Setting, error) {
	logrus.Info("UpsertSetting")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchId, err := primitive.ObjectIDFromHex(form.BranchId)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
		Upsert:         boolPtr(true),
	}

	data := entities.Setting{}
	err = entity.settingRepo.FindOneAndUpdate(ctx, bson.M{"branchId": branchId}, bson.M{
		"$set": bson.M{
			"branchId":       branchId,
			"receiptFooter":  form.ReceiptFooter,
			"companyName":    form.CompanyName,
			"companyAddress": form.CompanyAddress,
			"companyPhone":   form.CompanyPhone,
			"companyTaxId":   form.CompanyTaxId,
			"logoUrl":        form.LogoUrl,
			"showCredit":     form.ShowCredit,
			"promptPayId":    form.PromptPayId,
			"updatedBy":      form.UpdatedBy,
			"updatedDate":    time.Now(),
		},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func boolPtr(b bool) *bool {
	return &b
}
