package catagory

import (
	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
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
		middlewares.RequireSession(repository.Auth),
		usecase.GetCategories(repository.Category),
	)

	productRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateCategory(repository.Category),
	)

	productRoute.GET("/:categoryId",
		middlewares.RequireSession(repository.Auth),
		usecase.GetCategoryById(repository.Category),
	)

	productRoute.PUT("/:categoryId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateCategoryById(repository.Category),
	)

	productRoute.DELETE("/:categoryId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteCategoryById(repository.Category),
	)

	productRoute.PATCH("/:categoryId/default",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateDefaultCategoryById(repository.Category),
	)
}
