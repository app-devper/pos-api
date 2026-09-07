package repositories

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"pos/db"
)

const sessionPrefix = "session:"

type sessionEntity struct {
	rdb *redis.Client
}

type ISession interface {
	GetSessionById(sessionId string) (string, error)
}

func NewSessionEntity(resource *db.Resource) ISession {
	entity := &sessionEntity{rdb: resource.RdDb}
	return entity
}

func (entity *sessionEntity) GetSessionById(sessionId string) (string, error) {
	logrus.Info("GetSessionById")
	raw, err := entity.rdb.Get(context.Background(), sessionPrefix+sessionId).Result()
	if err != nil {
		return "", err
	}
	return parseSessionUserId([]byte(raw))
}

func parseSessionUserId(raw []byte) (string, error) {
	var data struct {
		UserId string `json:"userId"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", err
	}
	return data.UserId, nil
}
