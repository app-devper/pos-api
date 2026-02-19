package credit_note

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/credit_note/usecase"
	"pos/middlewares"
)

func ApplyCreditNoteAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	cnRoute := route.Group("credit-notes")

	cnRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateCreditNote(repository.CreditNote, repository.Product, repository.Sequence),
	)

	cnRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetCreditNotes(repository.CreditNote),
	)

	cnRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetCreditNoteById(repository.CreditNote),
	)

	cnRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateCreditNoteById(repository.CreditNote),
	)

	cnRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteCreditNoteById(repository.CreditNote),
	)
}
