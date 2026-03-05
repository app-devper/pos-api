package report

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/report/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyReportAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	reportRoute := route.Group("reports")

	reportRoute.GET("/receipt/:orderId/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetReceiptPDF(repository.Order, repository.Setting),
	)

	reportRoute.GET("/tax-invoice/:orderId/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetTaxInvoicePDF(repository.Order, repository.Setting),
	)

	reportRoute.GET("/sales/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetSalesReportPDF(repository.Order, repository.Setting),
	)

	reportRoute.GET("/sales/excel",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetSalesReportExcel(repository.Order),
	)

	reportRoute.GET("/stocks/excel",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetStockReportExcel(repository.Product),
	)

	reportRoute.GET("/drug-label/:logId/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDrugLabelPDF(repository.DispensingLog, repository.Setting),
	)

	reportRoute.GET("/pharmacy/khy9",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY9PDF(repository.Receive, repository.Product, repository.Setting),
	)

	reportRoute.GET("/pharmacy/khy10",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY10PDF(repository.DispensingLog, repository.Product, repository.Setting),
	)

	reportRoute.GET("/pharmacy/khy11",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY11PDF(repository.DispensingLog, repository.Product, repository.Setting),
	)

	reportRoute.GET("/pharmacy/khy12",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY12PDF(repository.DispensingLog, repository.Product, repository.Setting),
	)

	reportRoute.GET("/pharmacy/khy13",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY13PDF(repository.DispensingLog, repository.Product, repository.Setting),
	)

	// KHY CSV exports
	reportRoute.GET("/pharmacy/khy9/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY9CSV(repository.Receive, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy10/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY10CSV(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy11/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY11CSV(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy12/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY12CSV(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy13/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY13CSV(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/product-history/:productId/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryPDF(repository.Product, repository.Setting),
	)

	reportRoute.GET("/product-history/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryRangePDF(repository.Product, repository.Setting),
	)

	reportRoute.GET("/customer-history/:customerCode/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetCustomerHistoryPDF(repository.CustomerHistory, repository.Customer, repository.Setting),
	)

	reportRoute.POST("/barcodes/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetBarcodePDF(repository.Product),
	)

	reportRoute.GET("/price-tags/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPriceTagPDF(repository.Product),
	)

	reportRoute.GET("/receives/summary/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetReceiveSummaryPDF(repository.Receive, repository.Setting),
	)

	reportRoute.GET("/prices/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPriceReportPDF(repository.Product, repository.Setting),
	)

	reportRoute.GET("/promptpay/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromptPayQR(repository.Setting),
	)

	reportRoute.GET("/promptpay/payload",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromptPayPayload(repository.Setting),
	)
}
