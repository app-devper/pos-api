package product

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/product/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
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
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateProduct(repository.Product),
	)

	productRoute.POST("/receive",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
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
		middlewares.RequireBranch(repository.Employee, repository.Branch),
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
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductStocksByProductId(repository.Product),
	)

	productRoute.POST("/stocks",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateProductStock(repository.Product),
	)

	productRoute.PUT("/stocks/:stockId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductStockById(repository.Product),
	)

	productRoute.DELETE("/stocks/:stockId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.RemoveProductStockById(repository.Product),
	)

	productRoute.PATCH("/stocks/:stockId/quantity",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
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
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateProductUnit(repository.Product),
	)

	productRoute.PUT("/units/:unitId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductUnitById(repository.Product),
	)

	productRoute.DELETE("/units/:unitId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
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
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateProductPrice(repository.Product),
	)

	productRoute.PUT("/prices/:priceId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductPriceById(repository.Product),
	)

	productRoute.DELETE("/prices/:priceId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.RemoveProductPriceById(repository.Product),
	)

	// Product History
	productRoute.GET("/:productId/histories",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryByProductId(repository.Product),
	)

	productRoute.GET("/histories",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryByDateRange(repository.Product),
	)

	// Product Lot
	productRoute.GET("/lots/expire-notify",
		usecase.GetProductLotsExpireNotify(repository.Product),
	)

}
