package repositories

import (
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type sequenceEntity struct {
	sequenceRepo *mongo.Collection
}

type sequenceDefinition struct {
	Prefix string
	Format int
	Type   string
}

type ISequence interface {
	NextSequence(field string) (*entities.Sequence, error)
	CreateSequence(field string, value int) (*entities.Sequence, error)
	GetSequenceByField(field string) (*entities.Sequence, error)
}

func NewSequenceEntity(resource *db.Resource) ISequence {
	sequenceRepo := resource.PosDb.Collection("sequences")
	createCollectionIndex(sequenceRepo, "sequences field", mongo.IndexModel{
		Keys:    bson.D{{Key: "field", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	entity := &sequenceEntity{sequenceRepo: sequenceRepo}
	return entity
}

func (entity *sequenceEntity) CreateSequence(field string, value int) (*entities.Sequence, error) {
	logrus.Info("CreateSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()

	definition := getSequenceDefinition(field)
	currentDate := getSequenceDate()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
		Upsert:         boolPointer(true),
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":    primitive.NewObjectID(),
			"field":  field,
			"value":  value,
			"prefix": definition.Prefix,
			"format": definition.Format,
			"type":   definition.Type,
			"date":   currentDate,
		},
	}

	var data entities.Sequence
	err := entity.sequenceRepo.FindOneAndUpdate(ctx, bson.M{"field": field}, update, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *sequenceEntity) GetSequenceByField(field string) (*entities.Sequence, error) {
	logrus.Info("GetSequenceByField")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.Sequence
	err := entity.sequenceRepo.FindOne(ctx, bson.M{"field": field}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *sequenceEntity) NextSequence(field string) (*entities.Sequence, error) {
	logrus.Info("NextSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()

	definition := getSequenceDefinition(field)
	currentDate := getSequenceDate()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
		Upsert:         boolPointer(true),
	}

	update := mongo.Pipeline{
		{{
			Key: "$set",
			Value: bson.M{
				"field":  field,
				"prefix": definition.Prefix,
				"format": definition.Format,
				"type":   definition.Type,
				"date":   buildNextSequenceDateExpression(definition.Type, currentDate),
				"value":  buildNextSequenceValueExpression(definition.Type, currentDate),
			},
		}},
	}

	var data entities.Sequence
	err := entity.sequenceRepo.FindOneAndUpdate(ctx, bson.M{"field": field}, update, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func getSequenceDate() string {
	location := utils.GetLocation()
	return time.Now().In(location).Format("20060102")
}

func getSequenceDefinition(field string) sequenceDefinition {
	definition := sequenceDefinition{
		Prefix: "",
		Format: 4,
		Type:   constant.NONE,
	}
	switch field {
	case constant.ORDER:
		definition.Prefix = "OD_"
		definition.Type = constant.DAILY
	case constant.RECEIVE:
		definition.Prefix = "RC_"
		definition.Type = constant.DAILY
	case constant.MEMBER:
		definition.Prefix = "MB_"
		definition.Type = constant.YEARLY
	case constant.PRODUCT:
		definition.Prefix = "PD_"
	}
	return definition
}

func buildNextSequenceValueExpression(sequenceType string, currentDate string) interface{} {
	currentValue := bson.M{"$ifNull": bson.A{"$value", 0}}
	switch sequenceType {
	case constant.DAILY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, currentDate}},
				bson.M{"$add": bson.A{currentValue, 1}},
				1,
			},
		}
	case constant.MONTHLY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$substrCP": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, 0, 6}}, currentDate[:6]}},
				bson.M{"$add": bson.A{currentValue, 1}},
				1,
			},
		}
	case constant.YEARLY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$substrCP": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, 0, 4}}, currentDate[:4]}},
				bson.M{"$add": bson.A{currentValue, 1}},
				1,
			},
		}
	default:
		return bson.M{"$add": bson.A{currentValue, 1}}
	}
}

func buildNextSequenceDateExpression(sequenceType string, currentDate string) interface{} {
	switch sequenceType {
	case constant.DAILY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, currentDate}},
				bson.M{"$ifNull": bson.A{"$date", currentDate}},
				currentDate,
			},
		}
	case constant.MONTHLY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$substrCP": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, 0, 6}}, currentDate[:6]}},
				bson.M{"$ifNull": bson.A{"$date", currentDate}},
				currentDate,
			},
		}
	case constant.YEARLY:
		return bson.M{
			"$cond": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$substrCP": bson.A{bson.M{"$ifNull": bson.A{"$date", ""}}, 0, 4}}, currentDate[:4]}},
				bson.M{"$ifNull": bson.A{"$date", currentDate}},
				currentDate,
			},
		}
	default:
		return bson.M{"$ifNull": bson.A{"$date", currentDate}}
	}
}

func boolPointer(value bool) *bool {
	return &value
}
