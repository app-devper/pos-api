package repository

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"pos/db"
)

type sessionEntity struct {
	rdb *redis.Client
}

type ISession interface {
	GetSessionById(sessionId string) (string, error)
}

func NewSessionEntity(resource *db.Resource) ISession {
	var entity ISession = &sessionEntity{rdb: resource.RdDB}
	return entity
}

func (entity *sessionEntity) GetSessionById(sessionId string) (string, error) {
	logrus.Info("GetSessionById")
	result, err := entity.rdb.Get(context.Background(), sessionId).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}
