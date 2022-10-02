package api

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain/repository"
	"pos/app/domain/usecase"
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
		middlewares.RequireSession(sessionEntity),
		usecase.GetCategories(categoryEntity),
	)

	productRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.CreateCategory(categoryEntity),
	)

	productRoute.GET("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.GetCategoryById(categoryEntity),
	)

	productRoute.PUT("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.UpdateCategoryById(categoryEntity),
	)

	productRoute.DELETE("/:categoryId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.DeleteCategoryById(categoryEntity),
	)

	productRoute.PATCH("/:categoryId/default",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(sessionEntity),
		usecase.UpdateDefaultCategoryById(categoryEntity),
	)
}
