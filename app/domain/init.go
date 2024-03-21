package domain

import (
	"pos/app/data/repositories"
	"pos/db"
)

type Repository struct {
	Session  repositories.ISession
	Sequence repositories.ISequence
	Category repositories.ICategory
	Order    repositories.IOrder
	Product  repositories.IProduct
	Customer repositories.ICustomer
	Supplier repositories.ISupplier
	Receive  repositories.IReceive
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Session:  repositories.NewSessionEntity(resource),
		Category: repositories.NewCategoryEntity(resource),
		Order:    repositories.NewOrderEntity(resource),
		Sequence: repositories.NewSequenceEntity(resource),
		Customer: repositories.NewCustomerEntity(resource),
		Product:  repositories.NewProductEntity(resource),
		Supplier: repositories.NewSupplierEntity(resource),
		Receive:  repositories.NewReceiveEntity(resource),
	}
}
