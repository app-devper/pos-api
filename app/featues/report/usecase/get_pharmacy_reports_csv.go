package usecase

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetKHY9CSV(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")

		receiveRange := request.GetReceiveRange{
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
			BranchId:  branchId,
		}
		receives, err := receiveEntity.GetReceives(receiveRange)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
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

		ctx.Header("Content-Type", "text/csv; charset=utf-8")
		ctx.Header("Content-Disposition", "attachment; filename=khy9-report.csv")
		// BOM for Excel UTF-8
		ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

		w := csv.NewWriter(ctx.Writer)
		w.Write([]string{"#", "Date", "Code", "Product", "Lot", "Qty", "Cost"})

		row := 1
		for _, recv := range receives {
			for _, item := range recv.Items {
				product, ok := productMap[item.ProductId.Hex()]
				if !ok || product.DrugInfo == nil {
					continue
				}
				w.Write([]string{
					fmt.Sprintf("%d", row),
					recv.CreatedDate.Format("02/01/2006"),
					recv.Code,
					product.Name,
					item.LotNumber,
					fmt.Sprintf("%d", item.Quantity),
					fmt.Sprintf("%.2f", item.CostPrice),
				})
				row++
			}
		}
		w.Flush()
	}
}

func generateDispensingCSV(ctx *gin.Context, dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, branchId string, req pharmacyReportRange, drugType string, filename string) {
	logs, err := dispensingEntity.GetDispensingLogsByDateRange(branchId, req.StartDate, req.EndDate)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
		return
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

	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", filename))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(ctx.Writer)
	w.Write([]string{"#", "Date", "Drug Name", "Generic Name", "Qty", "Pharmacist", "License"})

	row := 1
	for _, log := range logs {
		for _, item := range log.Items {
			product, ok := logProductMap[item.ProductId.Hex()]
			if !ok || product.DrugInfo == nil || product.DrugInfo.DrugType != drugType {
				continue
			}
			w.Write([]string{
				fmt.Sprintf("%d", row),
				log.CreatedDate.Format("02/01/2006"),
				item.ProductName,
				item.GenericName,
				fmt.Sprintf("%d", item.Quantity),
				log.PharmacistName,
				log.LicenseNo,
			})
			row++
		}
	}
	w.Flush()
}

func GetKHY10CSV(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingCSV(ctx, dispensingEntity, productEntity, branchId, req, "CONTROLLED", "khy10-report")
	}
}

func GetKHY11CSV(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingCSV(ctx, dispensingEntity, productEntity, branchId, req, "DANGEROUS", "khy11-report")
	}
}

func GetKHY12CSV(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingCSV(ctx, dispensingEntity, productEntity, branchId, req, "PSYCHO", "khy12-report")
	}
}

func GetKHY13CSV(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingCSV(ctx, dispensingEntity, productEntity, branchId, req, "NARCOTIC", "khy13-report")
	}
}
