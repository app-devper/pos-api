package app

import (
	"os"
	"pos/app/domain"
	"pos/app/featues/billing"
	"pos/app/featues/branch"
	"pos/app/featues/catagory"
	"pos/app/featues/credit_note"
	"pos/app/featues/customer"
	"pos/app/featues/customer_history"
	"pos/app/featues/dashboard"
	"pos/app/featues/delivery_order"
	"pos/app/featues/dispensing"
	"pos/app/featues/employee"
	"pos/app/featues/order"
	"pos/app/featues/patient"
	"pos/app/featues/product"
	"pos/app/featues/promotion"
	"pos/app/featues/purchase_order"
	"pos/app/featues/quotation"
	"pos/app/featues/receive"
	"pos/app/featues/report"
	"pos/app/featues/setting"
	"pos/app/featues/stock_transfer"
	"pos/app/featues/supplier"
	"pos/db"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Routes struct {
}

func (app Routes) StartGin() {
	r := gin.New()

	err := r.SetTrustedProxies(nil)
	if err != nil {
		logrus.Error(err)
	}

	r.Use(gin.Logger())
	r.Use(middlewares.NewRecovery())
	r.Use(middlewares.NewCors([]string{"*"}))

	resource, err := db.InitResource()
	if err != nil {
		logrus.Error(err)
	}
	defer resource.Close()

	publicRoute := r.Group("/api/pos/v1")

	repository := domain.InitRepository(resource)

	product.ApplyProductAPI(publicRoute, repository)
	order.ApplyOrderAPI(publicRoute, repository)
	catagory.ApplyCategoryAPI(publicRoute, repository)
	customer.ApplyCustomerAPI(publicRoute, repository)
	supplier.ApplySupplierAPI(publicRoute, repository)
	receive.ApplyReceiveAPI(publicRoute, repository)
	branch.ApplyBranchAPI(publicRoute, repository)
	employee.ApplyEmployeeAPI(publicRoute, repository)
	dashboard.ApplyDashboardAPI(publicRoute, repository)
	report.ApplyReportAPI(publicRoute, repository)
	setting.ApplySettingAPI(publicRoute, repository)
	purchase_order.ApplyPurchaseOrderAPI(publicRoute, repository)
	delivery_order.ApplyDeliveryOrderAPI(publicRoute, repository)
	credit_note.ApplyCreditNoteAPI(publicRoute, repository)
	billing.ApplyBillingAPI(publicRoute, repository)
	quotation.ApplyQuotationAPI(publicRoute, repository)
	promotion.ApplyPromotionAPI(publicRoute, repository)
	customer_history.ApplyCustomerHistoryAPI(publicRoute, repository)
	patient.ApplyPatientAPI(publicRoute, repository)
	dispensing.ApplyDispensingAPI(publicRoute, repository)
	stock_transfer.ApplyStockTransferAPI(publicRoute, repository)

	r.NoRoute(middlewares.NoRoute())

	err = r.Run(":" + os.Getenv("PORT"))
	if err != nil {
		logrus.Error(err)
	}
}
