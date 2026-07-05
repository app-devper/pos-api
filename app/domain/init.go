package domain

import (
	"pos/app/data/repositories"
	"pos/db"
)

type Repository struct {
	Session         repositories.ISession
	Sequence        repositories.ISequence
	Category        repositories.ICategory
	Order           repositories.IOrder
	Product         repositories.IProduct
	ProductStock    repositories.IProductStock
	Customer        repositories.ICustomer
	Supplier        repositories.ISupplier
	Receive         repositories.IReceive
	Branch          repositories.IBranch
	Employee        repositories.IEmployee
	Setting         repositories.ISetting
	Promotion       repositories.IPromotion
	CustomerHistory repositories.ICustomerHistory
	Patient         repositories.IPatient
	StockTransfer   repositories.IStockTransfer
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Session:         repositories.NewSessionEntity(resource),
		Category:        repositories.NewCategoryEntity(resource),
		Order:           repositories.NewOrderEntity(resource),
		Sequence:        repositories.NewSequenceEntity(resource),
		Customer:        repositories.NewCustomerEntity(resource),
		Product:         repositories.NewProductEntity(resource),
		ProductStock:    repositories.NewProductStockEntity(resource),
		Supplier:        repositories.NewSupplierEntity(resource),
		Receive:         repositories.NewReceiveEntity(resource),
		Branch:          repositories.NewBranchEntity(resource),
		Employee:        repositories.NewEmployeeEntity(resource),
		Setting:         repositories.NewSettingEntity(resource),
		Promotion:       repositories.NewPromotionEntity(resource),
		CustomerHistory: repositories.NewCustomerHistoryEntity(resource),
		Patient:         repositories.NewPatientEntity(resource),
		StockTransfer:   repositories.NewStockTransferEntity(resource),
	}
}
