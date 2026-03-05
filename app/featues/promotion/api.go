package promotion

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/promotion/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyPromotionAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	promoRoute := route.Group("promotions")

	promoRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreatePromotion(repository.Promotion),
	)

	promoRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromotions(repository.Promotion),
	)

	promoRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPromotionById(repository.Promotion),
	)

	promoRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdatePromotionById(repository.Promotion),
	)

	promoRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.DeletePromotionById(repository.Promotion),
	)

	promoRoute.POST("/apply",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.ApplyPromotion(repository.Promotion),
	)
}
