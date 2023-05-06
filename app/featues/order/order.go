package order

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain/repository"
	usecase2 "pos/app/featues/order/usecase"
	"pos/middlewares"
)

func ApplyOrderAPI(
	app *gin.RouterGroup,
	sessionEntity repository.ISession,
	orderEntity repository.IOrder,
	productEntity repository.IProduct,
) {
	orderRoute := app.Group("order")

	orderRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.CreateOrder(orderEntity, productEntity),
	)

	orderRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetOrdersRange(orderEntity),
	)

	orderRoute.GET("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetOrderById(orderEntity),
	)

	orderRoute.DELETE("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.DeleteOrderById(orderEntity, productEntity),
	)

	orderRoute.GET("/:orderId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.UpdateTotalCostById(orderEntity, productEntity),
	)

	orderRoute.GET("/item",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetOrderItemRange(orderEntity),
	)

	orderRoute.GET("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetOrderItemById(orderEntity),
	)

	orderRoute.DELETE("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.DeleteOrderItemById(orderEntity, productEntity),
	)

	orderRoute.GET("/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.GetOrderItemByProductId(orderEntity),
	)

	orderRoute.DELETE("/:orderId/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.DeleteOrderItemByOrderProductId(orderEntity, productEntity),
	)

}
