package receive

import (
	"pos/app/domain"
	"pos/app/featues/receive/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyReceiveAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	receiveRoute := route.Group("receives")

	receiveRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateReceive(repository.Receive, repository.Sequence, repository.Product),
	)

	receiveRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetReceivesRange(repository.Receive),
	)

	receiveRoute.GET("/:receiveId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetReceiveById(repository.Receive),
	)

	receiveRoute.PUT("/:receiveId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateReceiveById(repository.Receive, repository.Product),
	)

	receiveRoute.DELETE("/:receiveId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteReceiveById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/total-cost",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateReceiveTotalCostById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/items",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateReceiveItemsById(repository.Receive, repository.Product),
	)

	receiveRoute.PATCH("/:receiveId/import",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.ImportReceiveToStockWithReconciliation(repository.Receive, repository.Product, repository.Order, repository.ProductStock),
	)

}
