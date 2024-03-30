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

	// Product
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

	productRoute.POST("/receive",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateProductReceive(repository.Product, repository.Receive),
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

	productRoute.DELETE("/:productId/sold-first",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.ClearQuantitySoldFirstById(repository.Product),
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

	// Product Stock
	productRoute.GET("/:productId/stocks",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductStocksByProductId(repository.Product),
	)

	productRoute.POST("/stocks",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateProductStock(repository.Product),
	)

	productRoute.PUT("/stocks/:stockId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductStockById(repository.Product),
	)

	productRoute.DELETE("/stocks/:stockId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.RemoveProductStockById(repository.Product),
	)

	productRoute.PATCH("/stocks/:stockId/quantity",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductStockQuantityById(repository.Product),
	)

	productRoute.PATCH("/stocks/sequence",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductStockSequence(repository.Product),
	)

	// Product Unit
	productRoute.POST("/units",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateProductUnit(repository.Product),
	)

	productRoute.PUT("/units/:unitId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductUnitById(repository.Product),
	)

	productRoute.DELETE("/units/:unitId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.RemoveProductUnitById(repository.Product),
	)

	productRoute.GET("/:productId/units",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductUnitsByProductId(repository.Product),
	)

	// Product Price
	productRoute.GET("/:productId/prices",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetProductPricesByProductId(repository.Product),
	)

	productRoute.POST("/prices",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateProductPrice(repository.Product),
	)

	productRoute.PUT("/prices/:priceId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateProductPriceById(repository.Product),
	)

	productRoute.DELETE("/prices/:priceId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.RemoveProductPriceById(repository.Product),
	)

	// Product Lot
	productRoute.GET("/lots/expire-notify",
		usecase.GetProductLotsExpireNotify(repository.Product),
	)

}
