package promotion

import (
	"pos/app/domain"
	"pos/app/featues/promotion/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyPromotionAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	promoRoute := route.Group("promotions")
	branchAdmin := policies.BranchAdmin.On(promoRoute)
	staff := policies.Staff.On(promoRoute)

	branchAdmin.POST("",
		usecase.CreatePromotion(repository.Promotion),
	)

	staff.GET("",
		usecase.GetPromotions(repository.Promotion),
	)

	staff.GET("/:id",
		usecase.GetPromotionById(repository.Promotion),
	)

	branchAdmin.PUT("/:id",
		usecase.UpdatePromotionById(repository.Promotion),
	)

	branchAdmin.DELETE("/:id",
		usecase.DeletePromotionById(repository.Promotion),
	)

	staff.POST("/apply",
		usecase.ApplyPromotion(repository.Promotion),
	)
}
