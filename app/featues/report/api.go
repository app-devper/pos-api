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

	reportRoute.GET("/pharmacy/khy9/data",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY9Data(repository.Receive, repository.Product, repository.Supplier),
	)

	reportRoute.GET("/pharmacy/khy10/data",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY10Data(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy11/data",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY11Data(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy12/data",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY12Data(repository.DispensingLog, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy13/data",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY13Data(repository.DispensingLog, repository.Product),
	)

	// KHY CSV exports
	reportRoute.GET("/pharmacy/khy9/csv",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetKHY9CSV(repository.Receive, repository.Product, repository.Supplier),
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

	reportRoute.POST("/barcodes/pdf",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetBarcodePDF(repository.Product),
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
