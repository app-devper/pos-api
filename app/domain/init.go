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
	Customer        repositories.ICustomer
	Supplier        repositories.ISupplier
	Receive         repositories.IReceive
	Branch          repositories.IBranch
	Employee        repositories.IEmployee
	Setting         repositories.ISetting
	PurchaseOrder   repositories.IPurchaseOrder
	DeliveryOrder   repositories.IDeliveryOrder
	CreditNote      repositories.ICreditNote
	Billing         repositories.IBilling
	Quotation       repositories.IQuotation
	Promotion       repositories.IPromotion
	CustomerHistory repositories.ICustomerHistory
	Patient         repositories.IPatient
	DispensingLog   repositories.IDispensingLog
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
		Supplier:        repositories.NewSupplierEntity(resource),
		Receive:         repositories.NewReceiveEntity(resource),
		Branch:          repositories.NewBranchEntity(resource),
		Employee:        repositories.NewEmployeeEntity(resource),
		Setting:         repositories.NewSettingEntity(resource),
		PurchaseOrder:   repositories.NewPurchaseOrderEntity(resource),
		DeliveryOrder:   repositories.NewDeliveryOrderEntity(resource),
		CreditNote:      repositories.NewCreditNoteEntity(resource),
		Billing:         repositories.NewBillingEntity(resource),
		Quotation:       repositories.NewQuotationEntity(resource),
		Promotion:       repositories.NewPromotionEntity(resource),
		CustomerHistory: repositories.NewCustomerHistoryEntity(resource),
		Patient:         repositories.NewPatientEntity(resource),
		DispensingLog:   repositories.NewDispensingLogEntity(resource),
		StockTransfer:   repositories.NewStockTransferEntity(resource),
	}
}
