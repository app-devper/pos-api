package catagory

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain/repository"
	usecase2 "pos/app/featues/catagory/usecase"
	"pos/middlewares"
)

func ApplyCategoryAPI(
	app *gin.RouterGroup,
	sessionEntity repository.ISession,
	categoryEntity repository.ICategory,
) {
	productRoute := app.Group("category")

	productRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetCategories(categoryEntity),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.CreateCategory(categoryEntity),
	)

	productRoute.GET("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.GetCategoryById(categoryEntity),
	)

	productRoute.PUT("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.UpdateCategoryById(categoryEntity),
	)

	productRoute.DELETE("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.DeleteCategoryById(categoryEntity),
	)

	productRoute.PATCH("/:categoryId/default",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase2.UpdateDefaultCategoryById(categoryEntity),
	)
}
