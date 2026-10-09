package domain

import (
	"pos/app/data/repositories"
	"pos/db"

	"github.com/app-devper/um-api/sessionclient/ginauth"
)

type Repository struct {
	// Auth verifies UM tokens and sessions; set by the app at startup.
	Auth            *ginauth.Auth
	Sequence        repositories.ISequence
	Category        repositories.ICategory
	Order           repositories.IOrder
	OrderAnalytics  repositories.IOrderAnalytics
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
	StockAdjustment repositories.IStockAdjustment
	StockCount      repositories.IStockCount
	ProductReturn   repositories.IProductReturn
}

func InitRepository(resource *db.Resource) *Repository {
	return &Repository{
		Category:        repositories.NewCategoryEntity(resource),
		Order:           repositories.NewOrderEntity(resource),
		OrderAnalytics:  repositories.NewOrderAnalyticsEntity(resource),
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
		StockAdjustment: repositories.NewStockAdjustmentEntity(resource),
		StockCount:      repositories.NewStockCountEntity(resource),
		ProductReturn:   repositories.NewProductReturnEntity(resource),
	}
}
