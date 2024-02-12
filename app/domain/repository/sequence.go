package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/constant"
	"pos/app/domain/model"
	"pos/db"
	"time"
)

type sequenceEntity struct {
	sequenceRepo *mongo.Collection
}

type ISequence interface {
	NextSequence(field string) (*model.Sequence, error)
	CreateSequence(field string, value int) (*model.Sequence, error)
	GetSequenceByField(field string) (*model.Sequence, error)
}

func NewSequenceEntity(resource *db.Resource) ISequence {
	sequenceRepo := resource.PosDb.Collection("sequences")
	entity := &sequenceEntity{sequenceRepo: sequenceRepo}
	return entity
}

func (entity *sequenceEntity) CreateSequence(field string, value int) (*model.Sequence, error) {
	logrus.Info("CreateSequence")
	data, _ := entity.GetSequenceByField(field)
	if data != nil {
		return data, nil
	} else {
		ctx, cancel := utils.InitContext()
		defer cancel()
		data := model.Sequence{}
		data.Id = primitive.NewObjectID()
		data.Field = field
		data.Value = value
		data.Format = 4
		if field == constant.ORDER {
			data.Prefix = "OD_"
			data.Type = constant.DAILY
		} else if field == constant.RECEIVE {
			data.Prefix = "RC_"
			data.Type = constant.DAILY
		} else if field == constant.MEMBER {
			data.Prefix = "MB_"
			data.Type = constant.YEARLY
		} else if field == constant.PRODUCT {
			data.Prefix = "PD_"
			data.Type = constant.NONE
		} else {
			data.Prefix = ""
			data.Type = constant.NONE
		}
		data.Date = getSequenceDate()
		_, err := entity.sequenceRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}

func (entity *sequenceEntity) GetSequenceByField(field string) (*model.Sequence, error) {
	logrus.Info("GetSequenceByField")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Sequence
	err := entity.sequenceRepo.FindOne(ctx, bson.M{"field": field}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *sequenceEntity) NextSequence(field string) (*model.Sequence, error) {
	logrus.Info("NextSequence")
	data, _ := entity.GetSequenceByField(field)
	if data != nil {
		ctx, cancel := utils.InitContext()
		defer cancel()
		var date = getSequenceDate()
		if data.Type == constant.NONE {
			data.Value = data.Value + 1
		} else if data.Type == constant.DAILY {
			if date == data.Date {
				data.Value = data.Value + 1
			} else {
				data.Value = 1
				data.Date = date
			}
		} else if data.Type == constant.MONTHLY {
			if date[0:6] == data.Date[0:6] {
				data.Value = data.Value + 1
			} else {
				data.Value = 1
				data.Date = date
			}
		} else if data.Type == constant.YEARLY {
			if date[0:4] == data.Date[0:4] {
				data.Value = data.Value + 1
			} else {
				data.Value = 1
				data.Date = date
			}
		}
		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err := entity.sequenceRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		data, err := entity.CreateSequence(field, 1)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func getSequenceDate() string {
	location := utils.GetLocation()
	return time.Now().In(location).Format("20060102")
}
