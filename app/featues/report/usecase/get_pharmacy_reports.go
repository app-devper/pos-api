package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
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

func GetKHY9PDF(receiveEntity repositories.IReceive, productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
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

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "Pharmacy"
		if setting != nil && setting.CompanyName != "" {
			companyName = setting.CompanyName
		}

		doc := pdf.NewPDF()
		doc.AddPage()
		pdf.AddHeader(doc, companyName, "", "", "KHY.9 - Drug Purchase Record")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Period: %s - %s", req.StartDate.Format("02/01/2006"), req.EndDate.Format("02/01/2006")), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Date", "Code", "Product", "Lot", "Qty", "Cost"}
		widths := []float64{10, 25, 25, 50, 25, 20, 35}
		aligns := []string{"C", "L", "L", "L", "L", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

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

		row := 1
		for _, recv := range receives {
			for _, item := range recv.Items {
				product, ok := productMap[item.ProductId.Hex()]
				if !ok || product.DrugInfo == nil {
					continue
				}
				pdf.AddTableRow(doc, []string{
					fmt.Sprintf("%d", row),
					recv.CreatedDate.Format("02/01/2006"),
					recv.Code,
					product.Name,
					item.LotNumber,
					fmt.Sprintf("%d", item.Quantity),
					fmt.Sprintf("%.2f", item.CostPrice),
				}, widths, aligns)
				row++
			}
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=khy9-report.pdf")
		doc.Output(ctx.Writer)
	}
}

func GetKHY10PDF(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.10 - Specially Controlled Drug Sales Record", "CONTROLLED")
	}
}

func GetKHY11PDF(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.11 - Dangerous Drug Sales Record", "DANGEROUS")
	}
}

func generateDispensingReport(ctx *gin.Context, dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, settingEntity repositories.ISetting, branchId string, req pharmacyReportRange, title string, drugType string) {
	logs, err := dispensingEntity.GetDispensingLogsByDateRange(branchId, req.StartDate, req.EndDate)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
		return
	}

	setting, _ := settingEntity.GetSettingByBranchId(branchId)
	companyName := "Pharmacy"
	if setting != nil && setting.CompanyName != "" {
		companyName = setting.CompanyName
	}

	doc := pdf.NewPDF()
	doc.AddPage()
	pdf.AddHeader(doc, companyName, "", "", title)
	doc.SetFont("Arial", "", 9)
	doc.CellFormat(0, 5, fmt.Sprintf("Period: %s - %s", req.StartDate.Format("02/01/2006"), req.EndDate.Format("02/01/2006")), "", 1, "C", false, 0, "")
	doc.Ln(3)

	headers := []string{"#", "Date", "Drug Name", "Generic Name", "Qty", "Pharmacist", "License"}
	widths := []float64{10, 25, 40, 40, 15, 30, 30}
	aligns := []string{"C", "L", "L", "L", "R", "L", "L"}
	pdf.AddTableHeader(doc, headers, widths)

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

	row := 1
	for _, log := range logs {
		for _, item := range log.Items {
			product, ok := logProductMap[item.ProductId.Hex()]
			if !ok || product.DrugInfo == nil || product.DrugInfo.DrugType != drugType {
				continue
			}
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", row),
				log.CreatedDate.Format("02/01/2006"),
				item.ProductName,
				item.GenericName,
				fmt.Sprintf("%d", item.Quantity),
				log.PharmacistName,
				log.LicenseNo,
			}, widths, aligns)
			row++
		}
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s-report.pdf", drugType))
	doc.Output(ctx.Writer)
}

func GetKHY12PDF(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.12 - Prescription Drug Sales Record", "PSYCHO")
	}
}

func GetKHY13PDF(dispensingEntity repositories.IDispensingLog, productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.13 - FDA-Mandated Drug Sales Report", "NARCOTIC")
	}
}
