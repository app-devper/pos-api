package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"strconv"
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
		pdf.AddHeader(doc, companyName, "", "", "KHY.9 - Drug Receiving Record")
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
					"",
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
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.10 - Dangerous Drug Sales Record", "DANGEROUS")
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
		generateDispensingReport(ctx, dispensingEntity, productEntity, settingEntity, branchId, req, "KHY.11 - Specially Controlled Drug Sales Record", "SPECIALLY_CONTROLLED")
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

func GetKHY12PDF(productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		generateExpireReport(ctx, productEntity, settingEntity, branchId, "KHY.12 - Expired Drug Report", true, 0)
	}
}

func GetKHY13PDF(productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		daysStr := ctx.DefaultQuery("days", "90")
		days, _ := strconv.Atoi(daysStr)
		if days <= 0 {
			days = 90
		}
		generateExpireReport(ctx, productEntity, settingEntity, branchId, fmt.Sprintf("KHY.13 - Near-Expiry Drug Report (%d days)", days), false, days)
	}
}

func generateExpireReport(ctx *gin.Context, productEntity repositories.IProduct, settingEntity repositories.ISetting, branchId string, title string, expired bool, days int) {
	setting, _ := settingEntity.GetSettingByBranchId(branchId)
	companyName := "Pharmacy"
	if setting != nil && setting.CompanyName != "" {
		companyName = setting.CompanyName
	}

	expireRange := request.GetProductLotsExpireRange{
		StartDate: time.Now().AddDate(-1, 0, 0),
		EndDate:   time.Now().AddDate(1, 0, 0),
	}
	lots, err := productEntity.GetProductLotsExpireNotify(expireRange)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
		return
	}

	doc := pdf.NewPDF()
	doc.AddPage()
	pdf.AddHeader(doc, companyName, "", "", title)
	doc.Ln(3)

	headers := []string{"#", "Product", "Lot Number", "Expire Date", "Qty", "Status"}
	widths := []float64{10, 55, 35, 30, 20, 40}
	aligns := []string{"C", "L", "L", "L", "R", "L"}
	pdf.AddTableHeader(doc, headers, widths)

	now := time.Now()
	threshold := now.AddDate(0, 0, days)
	row := 1

	for _, lot := range lots {
		include := false
		status := ""
		if expired && lot.ExpireDate.Before(now) {
			include = true
			status = "EXPIRED"
		} else if !expired && lot.ExpireDate.After(now) && lot.ExpireDate.Before(threshold) {
			include = true
			daysLeft := int(lot.ExpireDate.Sub(now).Hours() / 24)
			status = fmt.Sprintf("%d days left", daysLeft)
		}
		if !include {
			continue
		}

		if lot.Product.DrugInfo == nil {
			continue
		}
		productName := lot.Product.Name

		pdf.AddTableRow(doc, []string{
			fmt.Sprintf("%d", row),
			productName,
			lot.LotNumber,
			lot.ExpireDate.Format("02/01/2006"),
			fmt.Sprintf("%d", lot.Quantity),
			status,
		}, widths, aligns)
		row++
	}

	ctx.Header("Content-Type", "application/pdf")
	filename := "khy12"
	if !expired {
		filename = "khy13"
	}
	ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s-report.pdf", filename))
	doc.Output(ctx.Writer)
}
