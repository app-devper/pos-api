package product

import (
	"pos/app/domain"
	"pos/app/featues/product/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyProductAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {

	productRoute := route.Group("products")

	// Product
	productRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProducts(repository.Product),
	)

	productRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateProduct(repository.Product),
	)

	productRoute.POST("/receive",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateProductReceive(repository.Product),
	)

	productRoute.GET("/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductById(repository.Product, repository.ProductStock),
	)

	productRoute.PUT("/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateProductById(repository.Product, repository.ProductStock),
	)

	productRoute.DELETE("/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteProductById(repository.Product),
	)

	productRoute.DELETE("/:productId/sold-first",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.ClearQuantitySoldFirstById(repository.Product),
	)

	productRoute.GET("/serial-number/:serialNumber",
		middlewares.RequireSession(repository.Auth),
		usecase.GetProductBySerialNumber(repository.Product),
	)

	productRoute.GET("/serial-number",
		middlewares.RequireSession(repository.Auth),
		usecase.GenerateSerialNumber(repository.Sequence),
	)

	// Product Stock
	productRoute.GET("/:productId/stocks",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductStocksByProductId(repository.ProductStock),
	)

	productRoute.POST("/stocks",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateProductStock(repository.ProductStock),
	)

	productRoute.PUT("/stocks/:stockId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateProductStockById(repository.ProductStock, repository.Product),
	)

	productRoute.DELETE("/stocks/:stockId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.RemoveProductStockById(repository.ProductStock),
	)

	productRoute.PATCH("/stocks/:stockId/quantity",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateProductStockQuantityById(repository.ProductStock),
	)

	productRoute.PATCH("/stocks/sequence",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductStockSequence(repository.ProductStock),
	)

	// Product Unit
	productRoute.POST("/units",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateProductUnit(repository.Product),
	)

	productRoute.PUT("/units/:unitId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductUnitById(repository.Product, repository.ProductStock),
	)

	productRoute.DELETE("/units/:unitId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.RemoveProductUnitById(repository.Product),
	)

	productRoute.GET("/:productId/units",
		middlewares.RequireSession(repository.Auth),
		usecase.GetProductUnitsByProductId(repository.Product),
	)

	// Product Price
	productRoute.GET("/:productId/prices",
		middlewares.RequireSession(repository.Auth),
		usecase.GetProductPricesByProductId(repository.Product),
	)

	productRoute.POST("/prices",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateProductPrice(repository.Product),
	)

	productRoute.PUT("/prices/:priceId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateProductPriceById(repository.Product, repository.ProductStock),
	)

	productRoute.DELETE("/prices/:priceId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.RemoveProductPriceById(repository.Product),
	)

	// Product History
	productRoute.GET("/:productId/histories",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryByProductId(repository.ProductStock),
	)

	productRoute.GET("/histories",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductHistoryByDateRange(repository.ProductStock),
	)

	// Product Lot
	productRoute.GET("/lots",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetAllLots(repository.Product),
	)

	productRoute.GET("/lots/expire-notify",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetProductLotsExpireNotify(repository.ProductStock),
	)

	productRoute.GET("/lots/:lotId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetLotById(repository.Product),
	)

	productRoute.POST("/lots",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateLot(repository.Product),
	)

	productRoute.PUT("/lots/:lotId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateLotById(repository.Product),
	)

	productRoute.DELETE("/lots/:lotId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteLotById(repository.Product),
	)

	// CSV Import
	productRoute.POST("/import-csv",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.ImportCSV(repository.Product),
	)

	// Drug Interaction Check
	productRoute.POST("/drug-interaction-check",
		middlewares.RequireSession(repository.Auth),
		usecase.CheckDrugInteractions(repository.Product),
	)

}
