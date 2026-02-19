package stock_transfer

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/stock_transfer/usecase"
	"pos/middlewares"
)

func ApplyStockTransferAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	stRoute := route.Group("stock-transfers")

	stRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateStockTransfer(repository.StockTransfer, repository.Product, repository.Sequence),
	)

	stRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.GetStockTransfers(repository.StockTransfer),
	)

	stRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.GetStockTransferById(repository.StockTransfer),
	)

	stRoute.PATCH("/:id/approve",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.ApproveStockTransfer(repository.StockTransfer, repository.Product),
	)

	stRoute.PATCH("/:id/reject",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.RejectStockTransfer(repository.StockTransfer, repository.Product),
	)
}
