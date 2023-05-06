package order

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/order/usecase"
	"pos/middlewares"
)

func ApplyOrderAPI(
	app *gin.RouterGroup,
	repository *domain.Repository,
) {
	orderRoute := app.Group("order")

	orderRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateOrder(repository.Order, repository.Product),
	)

	orderRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetOrdersRange(repository.Order),
	)

	orderRoute.GET("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetOrderById(repository.Order),
	)

	orderRoute.DELETE("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderById(repository.Order, repository.Product),
	)

	orderRoute.GET("/:orderId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateTotalCostById(repository.Order, repository.Product),
	)

	orderRoute.GET("/item",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetOrderItemRange(repository.Order),
	)

	orderRoute.GET("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetOrderItemById(repository.Order),
	)

	orderRoute.DELETE("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemById(repository.Order, repository.Product),
	)

	orderRoute.GET("/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.GetOrderItemByProductId(repository.Order),
	)

	orderRoute.DELETE("/:orderId/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemByOrderProductId(repository.Order, repository.Product),
	)

}
