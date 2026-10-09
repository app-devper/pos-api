package order

import (
	"pos/app/domain"
	"pos/app/featues/order/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyOrderAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	orderRoute := route.Group("orders")

	orderRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateOrder(repository.Order),
	)

	orderRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrdersRange(repository.Order),
	)

	orderRoute.GET("/:orderId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderById(repository.Order),
	)

	orderRoute.DELETE("/:orderId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteOrderById(repository.Order, repository.Product),
	)

	orderRoute.DELETE("/:orderId/products/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteOrderItemByOrderProductId(repository.Order, repository.Product),
	)

	orderRoute.PATCH("/:orderId/customer-code",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateCustomerCodeOrderById(repository.Order),
	)

	orderRoute.GET("/customers/:customerCode",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrdersByCustomerCode(repository.Order),
	)

	orderRoute.GET("/items",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderItemRange(repository.Order),
	)

	orderRoute.GET("/items/:itemId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetOrderItemById(repository.Order),
	)

	orderRoute.DELETE("/items/:itemId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteOrderItemById(repository.Order, repository.Product),
	)

	orderRoute.GET("/items/products/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetOrderItemByProductId(repository.Order),
	)

	orderRoute.GET("/item-details/products/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetOrderItemDetailsByProductId(repository.Order),
	)

}
