package stock_transfer

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/stock_transfer/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockTransferAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	stRoute := route.Group("stock-transfers")

	stRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateStockTransfer(repository.StockTransfer, repository.Product, repository.Sequence),
	)

	stRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetStockTransfers(repository.StockTransfer),
	)

	stRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetStockTransferById(repository.StockTransfer),
	)

	stRoute.PATCH("/:id/approve",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.ApproveStockTransfer(repository.StockTransfer, repository.Product),
	)

	stRoute.PATCH("/:id/reject",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.RejectStockTransfer(repository.StockTransfer, repository.Product),
	)
}
