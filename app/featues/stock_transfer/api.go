package stock_transfer

import (
	"pos/app/domain"
	"pos/app/featues/stock_transfer/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockTransferAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	stRoute := route.Group("stock-transfers")
	branchAdmin := policies.BranchAdmin.On(stRoute)

	branchAdmin.POST("",
		usecase.CreateStockTransfer(repository.StockTransfer, repository.Product, repository.Sequence),
	)

	branchAdmin.GET("",
		usecase.GetStockTransfers(repository.StockTransfer),
	)

	branchAdmin.GET("/:id",
		usecase.GetStockTransferById(repository.StockTransfer),
	)

	branchAdmin.PATCH("/:id/approve",
		usecase.ApproveStockTransfer(repository.StockTransfer),
	)

	branchAdmin.PATCH("/:id/reject",
		usecase.RejectStockTransfer(repository.StockTransfer),
	)
}
