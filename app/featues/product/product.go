package product

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain/repository"
	usecase2 "pos/app/featues/product/usecase"
	"pos/middlewares"
)

func ApplyProductAPI(app *gin.RouterGroup,
	sessionEntity repository.ISession,
	productEntity repository.IProduct,
) {

	productRoute := app.Group("product")

	productRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetProducts(productEntity),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.CreateProduct(productEntity),
	)

	productRoute.GET("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetProductById(productEntity),
	)

	productRoute.PUT("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.UpdateProductById(productEntity),
	)

	productRoute.DELETE("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase2.DeleteProductById(productEntity),
	)

	productRoute.GET("/serial-number/:serialNumber",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetProductBySerialNumber(productEntity),
	)

}
