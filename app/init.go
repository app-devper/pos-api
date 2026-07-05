package app

import (
	"fmt"
	"net/http"
	"os"
	"pos/app/domain"
	"pos/app/domain/request"
	"pos/app/featues/branch"
	"pos/app/featues/catagory"
	"pos/app/featues/customer"
	"pos/app/featues/customer_history"
	"pos/app/featues/dashboard"
	"pos/app/featues/employee"
	"pos/app/featues/order"
	"pos/app/featues/patient"
	"pos/app/featues/product"
	"pos/app/featues/promotion"
	"pos/app/featues/receive"
	"pos/app/featues/report"
	"pos/app/featues/setting"
	"pos/app/featues/stock_transfer"
	"pos/app/featues/supplier"
	"pos/db"
	"pos/middlewares"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Routes struct {
}

func (app Routes) StartGin() error {
	configureGinMode()

	if err := validateStartupConfig(); err != nil {
		return err
	}

	r := gin.New()

	err := r.SetTrustedProxies(nil)
	if err != nil {
		logrus.Error(err)
	}

	r.Use(gin.Logger())
	r.Use(middlewares.NewRecovery())
	r.Use(middlewares.NewCors(getCorsOrigins()))

	resource, err := db.InitResource()
	if err != nil {
		return fmt.Errorf("init resource: %w", err)
	}
	defer resource.Close()

	publicRoute := r.Group("/api/pos/v1")

	repository := domain.InitRepository(resource)
	if shouldAutoInitDefaultBranch() {
		initDefaultBranch(repository)
	}

	product.ApplyProductAPI(publicRoute, repository)
	order.ApplyOrderAPI(publicRoute, repository)
	catagory.ApplyCategoryAPI(publicRoute, repository)
	customer.ApplyCustomerAPI(publicRoute, repository)
	supplier.ApplySupplierAPI(publicRoute, repository)
	receive.ApplyReceiveAPI(publicRoute, repository)
	branch.ApplyBranchAPI(publicRoute, repository)
	employee.ApplyEmployeeAPI(publicRoute, repository)
	dashboard.ApplyDashboardAPI(publicRoute, repository)
	report.ApplyReportAPI(publicRoute, repository)
	setting.ApplySettingAPI(publicRoute, repository)
	promotion.ApplyPromotionAPI(publicRoute, repository)
	customer_history.ApplyCustomerHistoryAPI(publicRoute, repository)
	patient.ApplyPatientAPI(publicRoute, repository)
	stock_transfer.ApplyStockTransferAPI(publicRoute, repository)

	r.NoRoute(middlewares.NoRoute())

	server := &http.Server{
		Addr:              ":" + getEnv("PORT", "8080"),
		Handler:           r,
		ReadTimeout:       getEnvDurationSeconds("HTTP_READ_TIMEOUT_SEC", 15),
		ReadHeaderTimeout: getEnvDurationSeconds("HTTP_READ_HEADER_TIMEOUT_SEC", 5),
		WriteTimeout:      getEnvDurationSeconds("HTTP_WRITE_TIMEOUT_SEC", 30),
		IdleTimeout:       getEnvDurationSeconds("HTTP_IDLE_TIMEOUT_SEC", 120),
		MaxHeaderBytes:    1 << 20,
	}

	err = server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("http server failed: %w", err)
	}

	return nil
}

func initDefaultBranch(repository *domain.Repository) {
	_, err := repository.Branch.GetBranchByCode("HQ")
	if err == nil {
		return
	}
	_, err = repository.Branch.CreateBranch(request.Branch{
		Name:      "สำนักงานใหญ่",
		Code:      "HQ",
		CreatedBy: "system",
	})
	if err != nil {
		logrus.Error("initDefaultBranch: failed to create default branch: ", err)
		return
	}
	logrus.Info("initDefaultBranch: created default branch 'สำนักงานใหญ่'")
}

func configureGinMode() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
		return
	}
	if isCloudRun() {
		gin.SetMode(gin.ReleaseMode)
	}
}

func shouldAutoInitDefaultBranch() bool {
	if value, ok := lookupBoolEnv("AUTO_INIT_DEFAULT_BRANCH"); ok {
		return value
	}
	return !isCloudRun()
}

func isCloudRun() bool {
	return os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvDurationSeconds(key string, fallbackSeconds int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Warn("invalid duration env, using fallback")
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func lookupBoolEnv(key string) (bool, bool) {
	value := os.Getenv(key)
	if value == "" {
		return false, false
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Warn("invalid bool env, ignoring")
		return false, false
	}
	return result, true
}

func getCorsOrigins() []string {
	if value := os.Getenv("CORS_ALLOWED_ORIGINS"); value != "" {
		return strings.Split(value, ",")
	}
	if isCloudRun() {
		logrus.Warn("CORS_ALLOWED_ORIGINS not set in production — defaulting to wildcard '*'. Set explicit origins for better security.")
	}
	return []string{"*"}
}

func validateStartupConfig() error {
	requiredKeys := []string{
		"SECRET_KEY",
		"CLIENT_ID",
		"SYSTEM",
		"MONGO_HOST",
		"MONGO_POS_DB_NAME",
		"REDIS_HOST",
	}

	for _, key := range requiredKeys {
		if os.Getenv(key) == "" {
			return fmt.Errorf("missing required env: %s", key)
		}
	}

	return nil
}
