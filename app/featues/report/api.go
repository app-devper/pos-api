package report

import (
	"pos/app/domain"
	"pos/app/featues/report/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyReportAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	reportRoute := route.Group("reports")
	branchAdmin := policies.BranchAdmin.On(reportRoute)
	staff := policies.Staff.On(reportRoute)

	branchAdmin.GET("/sales/excel",
		usecase.GetSalesReportExcel(repository.Order),
	)

	branchAdmin.GET("/stocks/excel",
		usecase.GetStockReportExcel(repository.ProductStock),
	)

	branchAdmin.GET("/pharmacy/khy9/data",
		usecase.GetKHY9Data(repository.Receive, repository.Product, repository.Supplier),
	)

	branchAdmin.GET("/pharmacy/khy10/data",
		usecase.GetKHY10Data(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy11/data",
		usecase.GetKHY11Data(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy12/data",
		usecase.GetKHY12Data(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy13/data",
		usecase.GetKHY13Data(repository.Order, repository.Product),
	)

	// KHY CSV exports
	branchAdmin.GET("/pharmacy/khy9/csv",
		usecase.GetKHY9CSV(repository.Receive, repository.Product, repository.Supplier),
	)

	branchAdmin.GET("/pharmacy/khy10/csv",
		usecase.GetKHY10CSV(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy11/csv",
		usecase.GetKHY11CSV(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy12/csv",
		usecase.GetKHY12CSV(repository.Order, repository.Product),
	)

	branchAdmin.GET("/pharmacy/khy13/csv",
		usecase.GetKHY13CSV(repository.Order, repository.Product),
	)

	staff.POST("/barcodes/pdf",
		usecase.GetBarcodePDF(repository.Product),
	)

	staff.GET("/promptpay/pdf",
		usecase.GetPromptPayQR(repository.Setting),
	)

	staff.GET("/promptpay/payload",
		usecase.GetPromptPayPayload(repository.Setting),
	)
}
