package catagory

import (
	"pos/app/domain"
	"pos/app/featues/catagory/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyCategoryAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	productRoute := route.Group("categories")
	shopAdmin := policies.ShopAdmin.On(productRoute)
	signedIn := policies.SignedIn.On(productRoute)

	signedIn.GET("",
		usecase.GetCategories(repository.Category),
	)

	shopAdmin.POST("",
		usecase.CreateCategory(repository.Category),
	)

	signedIn.GET("/:categoryId",
		usecase.GetCategoryById(repository.Category),
	)

	shopAdmin.PUT("/:categoryId",
		usecase.UpdateCategoryById(repository.Category),
	)

	shopAdmin.DELETE("/:categoryId",
		usecase.DeleteCategoryById(repository.Category),
	)

	shopAdmin.PATCH("/:categoryId/default",
		usecase.UpdateDefaultCategoryById(repository.Category),
	)
}
