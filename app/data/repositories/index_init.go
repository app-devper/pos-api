package repositories

import (
	"pos/app/core/utils"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type indexCreator interface {
	CreateIndex() (string, error)
}

func ensureCollectionIndex(entity indexCreator, name string) {
	if _, err := entity.CreateIndex(); err != nil {
		logrus.Errorf("failed to create %s index: %v", name, err)
	}
}

func createCollectionIndex(repo *mongo.Collection, name string, model mongo.IndexModel) {
	if repo == nil {
		logrus.Errorf("failed to create %s index: repo is nil", name)
		return
	}
	ctx, cancel := utils.InitContext()
	defer cancel()
	if _, err := repo.Indexes().CreateOne(ctx, model); err != nil {
		logrus.Errorf("failed to create %s index: %v", name, err)
	}
}
