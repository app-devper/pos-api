package product

import (
	"pos/app/domain"
	"pos/app/featues/product/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyProductAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)

	productRoute := route.Group("products")
	branchAdmin := policies.BranchAdmin.On(productRoute)
	signedIn := policies.SignedIn.On(productRoute)
	staff := policies.Staff.On(productRoute)

	// Product
	staff.GET("",
		usecase.GetProducts(repository.Product),
	)

	branchAdmin.POST("",
		usecase.CreateProduct(repository.Product),
	)

	branchAdmin.POST("/receive",
		usecase.CreateProductReceive(repository.Product),
	)

	staff.GET("/:productId",
		usecase.GetProductById(repository.Product, repository.ProductStock),
	)

	branchAdmin.PUT("/:productId",
		usecase.UpdateProductById(repository.Product, repository.ProductStock),
	)

	branchAdmin.DELETE("/:productId",
		usecase.DeleteProductById(repository.Product),
	)

	branchAdmin.DELETE("/:productId/sold-first",
		usecase.ClearQuantitySoldFirstById(repository.Product),
	)

	signedIn.GET("/serial-number/:serialNumber",
		usecase.GetProductBySerialNumber(repository.Product),
	)

	signedIn.GET("/serial-number",
		usecase.GenerateSerialNumber(repository.Sequence),
	)

	// Product Stock
	staff.GET("/:productId/stocks",
		usecase.GetProductStocksByProductId(repository.ProductStock),
	)

	branchAdmin.POST("/stocks",
		usecase.CreateProductStock(repository.Ledger),
	)

	branchAdmin.PUT("/stocks/:stockId",
		usecase.UpdateProductStockById(repository.ProductStock, repository.Product),
	)

	branchAdmin.DELETE("/stocks/:stockId",
		usecase.RemoveProductStockById(repository.ProductStock, repository.Ledger),
	)

	branchAdmin.PATCH("/stocks/:stockId/quantity",
		usecase.UpdateProductStockQuantityById(repository.ProductStock, repository.Ledger),
	)

	branchAdmin.PATCH("/stocks/sequence",
		usecase.UpdateProductStockSequence(repository.ProductStock),
	)

	// Product Unit
	branchAdmin.POST("/units",
		usecase.CreateProductUnit(repository.Product),
	)

	branchAdmin.PUT("/units/:unitId",
		usecase.UpdateProductUnitById(repository.Product, repository.ProductStock),
	)

	branchAdmin.DELETE("/units/:unitId",
		usecase.RemoveProductUnitById(repository.Product),
	)

	signedIn.GET("/:productId/units",
		usecase.GetProductUnitsByProductId(repository.Product),
	)

	// Product Price
	signedIn.GET("/:productId/prices",
		usecase.GetProductPricesByProductId(repository.Product),
	)

	branchAdmin.POST("/prices",
		usecase.CreateProductPrice(repository.Product),
	)

	branchAdmin.PUT("/prices/:priceId",
		usecase.UpdateProductPriceById(repository.Product, repository.ProductStock),
	)

	branchAdmin.DELETE("/prices/:priceId",
		usecase.RemoveProductPriceById(repository.Product),
	)

	// Product History
	staff.GET("/:productId/histories",
		usecase.GetProductHistoryByProductId(repository.ProductStock),
	)

	staff.GET("/histories",
		usecase.GetProductHistoryByDateRange(repository.ProductStock),
	)

	// Product Lot
	staff.GET("/lots",
		usecase.GetAllLots(repository.Product),
	)

	staff.GET("/lots/expire-notify",
		usecase.GetProductLotsExpireNotify(repository.ProductStock),
	)

	staff.GET("/lots/:lotId",
		usecase.GetLotById(repository.Product),
	)

	branchAdmin.POST("/lots",
		usecase.CreateLot(repository.Product),
	)

	branchAdmin.PUT("/lots/:lotId",
		usecase.UpdateLotById(repository.Product),
	)

	branchAdmin.DELETE("/lots/:lotId",
		usecase.DeleteLotById(repository.Product),
	)

	// CSV Import
	branchAdmin.POST("/import-csv",
		usecase.ImportCSV(repository.Product),
	)

	// Drug Interaction Check
	signedIn.POST("/drug-interaction-check",
		usecase.CheckDrugInteractions(repository.Product),
	)

}
