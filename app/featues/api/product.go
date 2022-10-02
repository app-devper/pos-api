package api

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain/repository"
	"pos/app/domain/usecase"
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
		usecase.GetProducts(productEntity),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateProduct(productEntity),
	)

	productRoute.GET("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetProductById(productEntity),
	)

	productRoute.PUT("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateProductById(productEntity),
	)

	productRoute.DELETE("/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteProductById(productEntity),
	)

	productRoute.GET("/serial-number/:serialNumber",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetProductBySerialNumber(productEntity),
	)

}
