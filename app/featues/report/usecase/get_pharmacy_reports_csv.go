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

func GetKHY9CSV(receiveEntity repositories.IReceive, productEntity repositories.IProduct, supplierEntity repositories.ISupplier) gin.HandlerFunc {
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
		supplierIdSet := make(map[string]struct{})
		for _, recv := range receives {
			if !recv.SupplierId.IsZero() {
				supplierIdSet[recv.SupplierId.Hex()] = struct{}{}
			}
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

		supplierMap := make(map[string]string)
		for sid := range supplierIdSet {
			if sup, err := supplierEntity.GetSupplierById(sid); err == nil && sup != nil {
				supplierMap[sid] = sup.Name
			}
		}

		ctx.Header("Content-Type", "text/csv; charset=utf-8")
		ctx.Header("Content-Disposition", "attachment; filename=khy9-report.csv")
		ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

		w := csv.NewWriter(ctx.Writer)
		w.Write([]string{"#", "วันที่", "เลขที่", "ชื่อยา", "ชื่อสามัญ", "ความแรง", "ล็อต", "วันหมดอายุ", "จำนวน", "หน่วย", "ต้นทุน", "ผู้จำหน่าย"})

		row := 1
		for _, recv := range receives {
			supName := supplierMap[recv.SupplierId.Hex()]
			for _, item := range recv.Items {
				product, ok := productMap[item.ProductId.Hex()]
				if !ok || product.DrugInfo == nil {
					continue
				}
				expStr := ""
				if !item.ExpireDate.IsZero() {
					expStr = item.ExpireDate.Format("02/01/2006")
				}
				w.Write([]string{
					fmt.Sprintf("%d", row),
					recv.CreatedDate.Format("02/01/2006"),
					recv.Code,
					product.Name,
					product.DrugInfo.GenericName,
					product.DrugInfo.Strength,
					item.LotNumber,
					expStr,
					fmt.Sprintf("%d", item.Quantity),
					product.Unit,
					fmt.Sprintf("%.2f", item.CostPrice),
					supName,
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
	w.Write([]string{"#", "วันที่", "ชื่อยา", "ชื่อสามัญ", "ความแรง", "รูปแบบยา", "จำนวน", "หน่วย", "วิธีใช้", "เภสัชกร", "เลขใบอนุญาต"})

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
				product.DrugInfo.Strength,
				product.DrugInfo.DosageForm,
				fmt.Sprintf("%d", item.Quantity),
				item.Unit,
				item.Dosage,
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
