package report

import (
	"pos/app/domain"
	"pos/app/featues/report/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyReportAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	reportRoute := route.Group("reports")

	reportRoute.GET("/sales/excel",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetSalesReportExcel(repository.Order),
	)

	reportRoute.GET("/stocks/excel",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockReportExcel(repository.ProductStock),
	)

	reportRoute.GET("/pharmacy/khy9/data",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY9Data(repository.Receive, repository.Product, repository.Supplier),
	)

	reportRoute.GET("/pharmacy/khy10/data",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY10Data(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy11/data",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY11Data(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy12/data",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY12Data(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy13/data",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY13Data(repository.Order, repository.Product),
	)

	// KHY CSV exports
	reportRoute.GET("/pharmacy/khy9/csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY9CSV(repository.Receive, repository.Product, repository.Supplier),
	)

	reportRoute.GET("/pharmacy/khy10/csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY10CSV(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy11/csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY11CSV(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy12/csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY12CSV(repository.Order, repository.Product),
	)

	reportRoute.GET("/pharmacy/khy13/csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetKHY13CSV(repository.Order, repository.Product),
	)

	reportRoute.POST("/barcodes/pdf",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetBarcodePDF(repository.Product),
	)

	reportRoute.GET("/promptpay/pdf",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromptPayQR(repository.Setting),
	)

	reportRoute.GET("/promptpay/payload",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromptPayPayload(repository.Setting),
	)
}
