package receive

import (
	"pos/app/domain"
	"pos/app/featues/receive/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyReceiveAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	receiveRoute := route.Group("receives")
	branchAdmin := policies.BranchAdmin.On(receiveRoute)

	branchAdmin.POST("",
		usecase.CreateReceive(repository.Receive, repository.Sequence, repository.Product),
	)

	branchAdmin.GET("",
		usecase.GetReceivesRange(repository.Receive),
	)

	branchAdmin.GET("/:receiveId",
		usecase.GetReceiveById(repository.Receive),
	)

	branchAdmin.PUT("/:receiveId",
		usecase.UpdateReceiveById(repository.Receive, repository.Product),
	)

	branchAdmin.DELETE("/:receiveId",
		usecase.DeleteReceiveById(repository.Receive),
	)

	branchAdmin.PATCH("/:receiveId/total-cost",
		usecase.UpdateReceiveTotalCostById(repository.Receive),
	)

	branchAdmin.PATCH("/:receiveId/items",
		usecase.UpdateReceiveItemsById(repository.Receive, repository.Product),
	)

	branchAdmin.PATCH("/:receiveId/import",
		usecase.ImportReceiveToStock(repository.Receive, repository.Ledger),
	)

}
