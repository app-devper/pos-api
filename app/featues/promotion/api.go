package promotion

import (
	"pos/app/domain"
	"pos/app/featues/promotion/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyPromotionAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	promoRoute := route.Group("promotions")

	promoRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreatePromotion(repository.Promotion),
	)

	promoRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromotions(repository.Promotion),
	)

	promoRoute.GET("/:id",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromotionById(repository.Promotion),
	)

	promoRoute.PUT("/:id",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdatePromotionById(repository.Promotion),
	)

	promoRoute.DELETE("/:id",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeletePromotionById(repository.Promotion),
	)

	promoRoute.POST("/apply",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.ApplyPromotion(repository.Promotion),
	)
}
