package delivery_order

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/delivery_order/usecase"
	"pos/middlewares"
)

func ApplyDeliveryOrderAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	doRoute := route.Group("delivery-orders")

	doRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateDeliveryOrder(repository.DeliveryOrder, repository.Sequence),
	)

	doRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDeliveryOrders(repository.DeliveryOrder),
	)

	doRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDeliveryOrderById(repository.DeliveryOrder),
	)

	doRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateDeliveryOrderById(repository.DeliveryOrder),
	)

	doRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteDeliveryOrderById(repository.DeliveryOrder),
	)
}
