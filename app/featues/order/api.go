package order

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/order/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyOrderAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	orderRoute := route.Group("orders")

	orderRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateOrder(repository.Order, repository.Product, repository.Sequence),
	)

	orderRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrdersRange(repository.Order),
	)

	orderRoute.GET("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderById(repository.Order),
	)

	orderRoute.DELETE("/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderById(repository.Order, repository.Product),
	)

	orderRoute.DELETE("/:orderId/products/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemByOrderProductId(repository.Order, repository.Product),
	)

	orderRoute.PATCH("/:orderId/customer-code",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateCustomerCodeOrderById(repository.Order),
	)

	orderRoute.GET("/customers/:customerCode",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrdersByCustomerCode(repository.Order),
	)

	orderRoute.GET("/items",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderItemRange(repository.Order),
	)

	orderRoute.GET("/items/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderItemById(repository.Order),
	)

	orderRoute.DELETE("/items/:itemId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteOrderItemById(repository.Order, repository.Product),
	)

	orderRoute.GET("/items/products/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.GetOrderItemByProductId(repository.Order),
	)

	orderRoute.GET("/item-details/products/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.GetOrderItemDetailsByProductId(repository.Order),
	)

}
