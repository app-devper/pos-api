package purchase_order

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/purchase_order/usecase"
	"pos/middlewares"
)

func ApplyPurchaseOrderAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	poRoute := route.Group("purchase-orders")

	poRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreatePurchaseOrder(repository.PurchaseOrder, repository.Sequence),
	)

	poRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPurchaseOrders(repository.PurchaseOrder),
	)

	poRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPurchaseOrderById(repository.PurchaseOrder),
	)

	poRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdatePurchaseOrderById(repository.PurchaseOrder),
	)

	poRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeletePurchaseOrderById(repository.PurchaseOrder),
	)
}
