package receive

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain"
	"pos/app/featues/receive/usecase"
	"pos/middlewares"
)

func ApplyReceiveAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	receiveRoute := route.Group("receives")

	receiveRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateReceive(repository.Receive, repository.Sequence),
	)

	receiveRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetReceivesRange(repository.Receive),
	)

	receiveRoute.GET("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetReceiveById(repository.Receive),
	)

	receiveRoute.PUT("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateReceiveById(repository.Receive),
	)

	receiveRoute.DELETE("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.DeleteReceiveById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateReceiveTotalCostById(repository.Receive),
	)

	receiveRoute.GET("/:receiveId/lots",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetReceiveItemByReceiveId(repository.Receive, repository.Product),
	)

	receiveRoute.DELETE("lots/:lotId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.DeleteReceiveItemByLotId(repository.Receive, repository.Product),
	)

}
