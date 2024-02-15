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

	productRoute := route.Group("products")

	productRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProducts(repository.Product),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateProduct(repository.Product, repository.Receive),
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

	productRoute.GET("/serial-number",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GenerateSerialNumber(repository.Sequence),
	)

	productRoute.POST("/prices",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateProductPrice(repository.Product),
	)

	productRoute.GET("/:productId/prices/customers/:customerId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductPriceByProductCustomerId(repository.Product),
	)

	productRoute.GET("/prices/customers/:customerId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductPriceDetailsByCustomerId(repository.Product),
	)

	productRoute.GET("/:productId/prices",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductPriceDetailsByProductId(repository.Product),
	)

	productRoute.GET("/:productId/lots",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductLotsByProductId(repository.Product),
	)

	productRoute.POST("/lots",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateProductLot(repository.Product),
	)

	productRoute.GET("/lots/expired",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductLotsExpired(repository.Product),
	)

	productRoute.GET("/lots/:lotId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductLotByLotId(repository.Product),
	)

	productRoute.PUT("/lots/:lotId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductLotByLotId(repository.Product),
	)

	productRoute.PATCH("/lots/:lotId/notify",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductLotNotifyByLotId(repository.Product),
	)

	productRoute.GET("/lots/expire-notify",
		usecase.GetProductLotsExpireNotify(repository.Product),
	)

}
