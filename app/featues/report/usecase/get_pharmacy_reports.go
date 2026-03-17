package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

type pharmacyReportRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

type pharmacyReportItem struct {
	Date           time.Time `json:"date"`
	Code           string    `json:"code,omitempty"`
	ProductName    string    `json:"productName"`
	GenericName    string    `json:"genericName,omitempty"`
	LotNumber      string    `json:"lotNumber,omitempty"`
	Quantity       int       `json:"quantity"`
	CostPrice      float64   `json:"costPrice,omitempty"`
	PharmacistName string    `json:"pharmacistName,omitempty"`
	LicenseNo      string    `json:"licenseNo,omitempty"`
	DrugType       string    `json:"drugType,omitempty"`
}

type pharmacyReportResponse struct {
	Key       string               `json:"key"`
	Title     string               `json:"title"`
	StartDate time.Time            `json:"startDate"`
	EndDate   time.Time            `json:"endDate"`
	Items     []pharmacyReportItem `json:"items"`
}

func getKHY9Items(receiveEntity repositories.IReceive, productEntity repositories.IProduct, branchId string, req pharmacyReportRange) ([]pharmacyReportItem, error) {
	receiveRange := request.GetReceiveRange{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		BranchId:  branchId,
	}
	receives, err := receiveEntity.GetReceives(receiveRange)
	if err != nil {
		return nil, err
	}

	productIdSet := make(map[string]struct{})
	for _, recv := range receives {
		for _, item := range recv.Items {
			productIdSet[item.ProductId.Hex()] = struct{}{}
		}
	}
	productIds := make([]string, 0, len(productIdSet))
	for id := range productIdSet {
		productIds = append(productIds, id)
	}
	productList, _ := productEntity.GetProductsByIds(productIds)
	productMap := make(map[string]*entities.Product, len(productList))
	for i := range productList {
		productMap[productList[i].Id.Hex()] = &productList[i]
	}

	items := make([]pharmacyReportItem, 0)
	for _, recv := range receives {
		for _, item := range recv.Items {
			product, ok := productMap[item.ProductId.Hex()]
			if !ok || product.DrugInfo == nil {
				continue
			}
			items = append(items, pharmacyReportItem{
				Date:        recv.CreatedDate,
				Code:        recv.Code,
				ProductName: product.Name,
				LotNumber:   item.LotNumber,
				Quantity:    item.Quantity,
				CostPrice:   item.CostPrice,
				DrugType:    product.DrugInfo.DrugType,
			})
		}
	}
	return items, nil
}

func getDispensingReportItems(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, branchId string, req pharmacyReportRange, drugType string) ([]pharmacyReportItem, error) {
	logs, err := dispensingEntity.GetDispensingLogsByDateRange(branchId, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	logProductIdSet := make(map[string]struct{})
	for _, log := range logs {
		for _, item := range log.Items {
			logProductIdSet[item.ProductId.Hex()] = struct{}{}
		}
	}
	logProductIds := make([]string, 0, len(logProductIdSet))
	for id := range logProductIdSet {
		logProductIds = append(logProductIds, id)
	}
	logProductList, _ := productEntity.GetProductsByIds(logProductIds)
	logProductMap := make(map[string]*entities.Product, len(logProductList))
	for i := range logProductList {
		logProductMap[logProductList[i].Id.Hex()] = &logProductList[i]
	}

	items := make([]pharmacyReportItem, 0)
	for _, log := range logs {
		for _, item := range log.Items {
			product, ok := logProductMap[item.ProductId.Hex()]
			if !ok || product.DrugInfo == nil || product.DrugInfo.DrugType != drugType {
				continue
			}
			items = append(items, pharmacyReportItem{
				Date:           log.CreatedDate,
				ProductName:    item.ProductName,
				GenericName:    item.GenericName,
				Quantity:       item.Quantity,
				PharmacistName: log.PharmacistName,
				LicenseNo:      log.LicenseNo,
				DrugType:       product.DrugInfo.DrugType,
			})
		}
	}
	return items, nil
}

func GetKHY9Data(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		items, err := getKHY9Items(receiveEntity, productEntity, branchId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, pharmacyReportResponse{Key: "khy9", Title: "KHY.9 - Drug Purchase Record", StartDate: req.StartDate, EndDate: req.EndDate, Items: items})
	}
}

func getDispensingReportDataHandler(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, key string, title string, drugType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		items, err := getDispensingReportItems(dispensingEntity, productEntity, branchId, req, drugType)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, pharmacyReportResponse{Key: key, Title: title, StartDate: req.StartDate, EndDate: req.EndDate, Items: items})
	}
}

func GetKHY10Data(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return getDispensingReportDataHandler(dispensingEntity, productEntity, "khy10", "KHY.10 - Specially Controlled Drug Sales Record", "CONTROLLED")
}

func GetKHY11Data(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return getDispensingReportDataHandler(dispensingEntity, productEntity, "khy11", "KHY.11 - Dangerous Drug Sales Record", "DANGEROUS")
}

func GetKHY12Data(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return getDispensingReportDataHandler(dispensingEntity, productEntity, "khy12", "KHY.12 - Prescription Drug Sales Record", "PSYCHO")
}

func GetKHY13Data(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return getDispensingReportDataHandler(dispensingEntity, productEntity, "khy13", "KHY.13 - FDA-Mandated Drug Sales Report", "NARCOTIC")
}
