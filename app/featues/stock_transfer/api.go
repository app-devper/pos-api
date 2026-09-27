package stock_transfer

import (
	"pos/app/domain"
	"pos/app/featues/stock_transfer/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyStockTransferAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	stRoute := route.Group("stock-transfers")

	stRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateStockTransfer(repository.StockTransfer, repository.Product, repository.Sequence),
	)

	stRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockTransfers(repository.StockTransfer),
	)

	stRoute.GET("/:id",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockTransferById(repository.StockTransfer),
	)

	stRoute.PATCH("/:id/approve",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.ApproveStockTransfer(repository.StockTransfer),
	)

	stRoute.PATCH("/:id/reject",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.RejectStockTransfer(repository.StockTransfer),
	)
}
