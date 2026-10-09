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
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	dashboardRoute := route.Group("dashboard")
	staff := policies.Staff.On(dashboardRoute)

	staff.GET("/summary",
		usecase.GetSummary(repository.OrderAnalytics),
	)

	staff.GET("/daily-chart",
		usecase.GetDailyChart(repository.OrderAnalytics),
	)

	staff.GET("/low-stock",
		usecase.GetLowStockProducts(repository.ProductStock),
	)

	staff.GET("/stock-report",
		usecase.GetStockReport(repository.ProductStock),
	)

	staff.GET("/monthly-chart",
		usecase.GetMonthlyChart(repository.OrderAnalytics),
	)

	staff.GET("/expiring",
		usecase.GetExpiringProducts(repository.ProductStock),
	)

	staff.GET("/refill-reminders",
		usecase.GetRefillReminders(),
	)

	staff.GET("/abc-analysis",
		usecase.GetABCAnalysis(repository.OrderAnalytics),
	)

	staff.GET("/dead-stock",
		usecase.GetDeadStockProducts(repository.ProductStock),
	)
}
