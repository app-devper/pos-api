package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetProductHistoryPDF(productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := ctx.GetString("BranchId")

		product, err := productEntity.GetProductById(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		histories, err := productEntity.GetProductHistoryByProductId(productId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "POS System"
		if setting != nil && setting.CompanyName != "" {
			companyName = setting.CompanyName
		}

		doc := pdf.NewPDF()
		doc.AddPage()
		pdf.AddHeader(doc, companyName, "", "", "Product History Report")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Product: %s (%s)", product.Name, product.SerialNumber), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Date", "Type", "Description", "Qty", "Cost", "Price", "Balance"}
		widths := []float64{10, 30, 25, 35, 18, 22, 22, 22}
		aligns := []string{"C", "L", "L", "L", "R", "R", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, h := range histories {
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				h.CreatedDate.Format("02/01/2006 15:04"),
				h.Type,
				h.Description,
				fmt.Sprintf("%d", h.Quantity),
				fmt.Sprintf("%.2f", h.CostPrice),
				fmt.Sprintf("%.2f", h.Price),
				fmt.Sprintf("%d", h.Balance),
			}, widths, aligns)
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=product-history-%s.pdf", product.SerialNumber))
		doc.Output(ctx.Writer)
	}
}

func GetProductHistoryRangePDF(productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetOrderRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")

		histories, err := productEntity.GetProductHistoryByDateRange(branchId, req.StartDate, req.EndDate)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "POS System"
		if setting != nil && setting.CompanyName != "" {
			companyName = setting.CompanyName
		}

		doc := pdf.NewPDF()
		doc.AddPage()
		pdf.AddHeader(doc, companyName, "", "", "Product History Report")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Period: %s - %s", req.StartDate.Format("02/01/2006"), req.EndDate.Format("02/01/2006")), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Date", "Type", "Unit", "Qty", "Cost", "Price", "Balance"}
		widths := []float64{10, 30, 30, 20, 20, 25, 25, 25}
		aligns := []string{"C", "L", "L", "L", "R", "R", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, h := range histories {
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				h.CreatedDate.Format("02/01/2006 15:04"),
				h.Type,
				h.Unit,
				fmt.Sprintf("%d", h.Quantity),
				fmt.Sprintf("%.2f", h.CostPrice),
				fmt.Sprintf("%.2f", h.Price),
				fmt.Sprintf("%d", h.Balance),
			}, widths, aligns)
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=product-history-report.pdf")
		doc.Output(ctx.Writer)
	}
}
