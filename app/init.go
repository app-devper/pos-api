package app

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"pos/app/domain/repository"
	"pos/app/featues/catagory"
	"pos/app/featues/order"
	"pos/app/featues/product"
	"pos/db"
	"pos/middlewares"
)

type Routes struct {
}

func (app Routes) StartGin() {
	r := gin.New()

	err := r.SetTrustedProxies(nil)
	if err != nil {
		logrus.Error(err)
	}

	r.Use(gin.Logger())
	r.Use(middlewares.NewRecovery())
	r.Use(middlewares.NewCors([]string{"*"}))

	resource, err := db.InitResource()
	if err != nil {
		logrus.Error(err)
	}
	defer resource.Close()

	publicRoute := r.Group("/api/pos/v1")

	sessionEntity := repository.NewSessionEntity(resource)
	productEntity := repository.NewProductEntity(resource)
	orderEntity := repository.NewOrderEntity(resource)
	categoryEntity := repository.NewCategoryEntity(resource)

	product.ApplyProductAPI(publicRoute, sessionEntity, productEntity)
	order.ApplyOrderAPI(publicRoute, sessionEntity, orderEntity, productEntity)
	catagory.ApplyCategoryAPI(publicRoute, sessionEntity, categoryEntity)

	r.NoRoute(middlewares.NoRoute())

	err = r.Run(":" + os.Getenv("PORT"))
	if err != nil {
		logrus.Error(err)
	}
}
