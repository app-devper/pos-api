package domain

import (
	"pos/app/domain/repository"
	"pos/db"
)

type Repository struct {
	Session  repository.ISession
	Sequence repository.ISequence
	Category repository.ICategory
	Order    repository.IOrder
	Product  repository.IProduct
	Customer repository.ICustomer
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Session:  repository.NewSessionEntity(resource),
		Category: repository.NewCategoryEntity(resource),
		Order:    repository.NewOrderEntity(resource),
		Sequence: repository.NewSequenceEntity(resource),
		Customer: repository.NewCustomerEntity(resource),
		Product:  repository.NewProductEntity(resource),
	}
}
