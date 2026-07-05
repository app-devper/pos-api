package dashboard

import (
	"pos/app/domain"
	"pos/app/featues/dashboard/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyDashboardAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	dashboardRoute := route.Group("dashboard")

	dashboardRoute.GET("/summary",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetSummary(repository.Order),
	)

	dashboardRoute.GET("/daily-chart",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDailyChart(repository.Order),
	)

	dashboardRoute.GET("/low-stock",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetLowStockProducts(repository.ProductStock),
	)

	dashboardRoute.GET("/stock-report",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetStockReport(repository.ProductStock),
	)

	dashboardRoute.GET("/monthly-chart",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetMonthlyChart(repository.Order),
	)

	dashboardRoute.GET("/expiring",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetExpiringProducts(repository.ProductStock),
	)

	dashboardRoute.GET("/refill-reminders",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetRefillReminders(),
	)

	dashboardRoute.GET("/abc-analysis",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetABCAnalysis(repository.Order),
	)

	dashboardRoute.GET("/dead-stock",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDeadStockProducts(repository.ProductStock),
	)
}
