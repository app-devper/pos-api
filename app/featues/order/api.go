package order

import (
	"pos/app/domain"
	"pos/app/featues/order/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyOrderAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	orderRoute := route.Group("orders")
	branchAdmin := policies.BranchAdmin.On(orderRoute)
	staff := policies.Staff.On(orderRoute)

	staff.POST("",
		usecase.CreateOrder(repository.Ledger),
	)

	staff.GET("",
		usecase.GetOrdersRange(repository.Order),
	)

	staff.GET("/:orderId",
		usecase.GetOrderById(repository.Order),
	)

	branchAdmin.DELETE("/:orderId",
		usecase.DeleteOrderById(repository.Order, repository.Product),
	)

	branchAdmin.DELETE("/:orderId/products/:productId",
		usecase.DeleteOrderItemByOrderProductId(repository.Order, repository.Product),
	)

	staff.PATCH("/:orderId/customer-code",
		usecase.UpdateCustomerCodeOrderById(repository.Order),
	)

	staff.GET("/customers/:customerCode",
		usecase.GetOrdersByCustomerCode(repository.Order),
	)

	staff.GET("/items",
		usecase.GetOrderItemRange(repository.Order),
	)

	staff.GET("/items/:itemId",
		usecase.GetOrderItemById(repository.Order),
	)

	branchAdmin.DELETE("/items/:itemId",
		usecase.DeleteOrderItemById(repository.Order, repository.Product),
	)

	branchAdmin.GET("/items/products/:productId",
		usecase.GetOrderItemByProductId(repository.Order),
	)

	branchAdmin.GET("/item-details/products/:productId",
		usecase.GetOrderItemDetailsByProductId(repository.Order),
	)

}
