package domain

import (
	repository2 "pos/app/data/repository"
	"pos/db"
)

type Repository struct {
	Session  repository2.ISession
	Sequence repository2.ISequence
	Category repository2.ICategory
	Order    repository2.IOrder
	Product  repository2.IProduct
	Customer repository2.ICustomer
	Supplier repository2.ISupplier
	Receive  repository2.IReceive
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Session:  repository2.NewSessionEntity(resource),
		Category: repository2.NewCategoryEntity(resource),
		Order:    repository2.NewOrderEntity(resource),
		Sequence: repository2.NewSequenceEntity(resource),
		Customer: repository2.NewCustomerEntity(resource),
		Product:  repository2.NewProductEntity(resource),
		Supplier: repository2.NewSupplierEntity(resource),
		Receive:  repository2.NewReceiveEntity(resource),
	}
}
