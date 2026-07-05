package catagory

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/catagory/usecase"
	"pos/middlewares"
)

func ApplyCategoryAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	productRoute := route.Group("categories")

	productRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetCategories(repository.Category),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateCategory(repository.Category),
	)

	productRoute.GET("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetCategoryById(repository.Category),
	)

	productRoute.PUT("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateCategoryById(repository.Category),
	)

	productRoute.DELETE("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.DeleteCategoryById(repository.Category),
	)

	productRoute.PATCH("/:categoryId/default",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateDefaultCategoryById(repository.Category),
	)
}
