package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// excludeArchivedParam opts the list out of archived Customers; without it
// the list is every Customer, archived ones included.
const excludeArchivedParam = "excludeArchived"

func GetCustomers(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		excludeArchived, ok := request.BoolQuery(ctx, excludeArchivedParam)
		if !ok {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, excludeArchivedParam+" must be true or false")
			return
		}
		result, err := customerEntity.GetCustomerAll(request.GetCustomers{ExcludeArchived: excludeArchived})
		if err != nil {
			logrus.WithError(err).Error("get customers failed")
			errcode.AbortLookup(ctx, err, errcode.CU_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
