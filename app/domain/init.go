package domain

import (
	"pos/app/domain/repository"
	"pos/db"
)

type Repository struct {
	Session  repository.ISession
	Category repository.ICategory
	Order    repository.IOrder
	Product  repository.IProduct
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Session:  repository.NewSessionEntity(resource),
		Category: repository.NewCategoryEntity(resource),
		Order:    repository.NewOrderEntity(resource),
		Product:  repository.NewProductEntity(resource),
	}
}
