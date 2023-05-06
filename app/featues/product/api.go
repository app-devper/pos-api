package product

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/product/usecase"
	"pos/middlewares"
)

func ApplyProductAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {

	productRoute := route.Group("product")

	productRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProducts(repository.Product),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateProduct(repository.Product),
	)

	productRoute.GET("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductById(repository.Product),
	)

	productRoute.PUT("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateProductById(repository.Product),
	)

	productRoute.DELETE("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteProductById(repository.Product),
	)

	productRoute.GET("/serial-number/:serialNumber",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductBySerialNumber(repository.Product),
	)

}
