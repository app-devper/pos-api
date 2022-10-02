package api

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain/repository"
	"pos/app/domain/usecase"
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
		usecase.CreateOrder(orderEntity, productEntity),
	)

	orderRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetOrdersRange(orderEntity),
	)

	orderRoute.GET("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetOrderById(orderEntity),
	)

	orderRoute.DELETE("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderById(orderEntity, productEntity),
	)

	orderRoute.GET("/:orderId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateTotalCostById(orderEntity, productEntity),
	)

	orderRoute.GET("/item",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetOrderItemRange(orderEntity),
	)

	orderRoute.GET("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetOrderItemById(orderEntity),
	)

	orderRoute.DELETE("/item/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemById(orderEntity, productEntity),
	)

	orderRoute.GET("/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.GetOrderItemByProductId(orderEntity),
	)

	orderRoute.DELETE("/:orderId/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemByOrderProductId(orderEntity, productEntity),
	)

}
