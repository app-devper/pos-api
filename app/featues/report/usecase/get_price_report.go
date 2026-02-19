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

func GetPriceReportPDF(productEntity repositories.IProduct, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		req := request.GetProduct{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			req = request.GetProduct{}
		}

		products, err := productEntity.GetProductAll(req)
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
		pdf.AddHeader(doc, companyName, "", "", "Price Report")
		doc.Ln(3)

		headers := []string{"#", "Serial No.", "Product Name", "Unit", "Cost", "Price"}
		widths := []float64{10, 30, 60, 25, 30, 35}
		aligns := []string{"C", "L", "L", "L", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, p := range products {
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				p.SerialNumber,
				p.Name,
				p.Unit,
				fmt.Sprintf("%.2f", p.CostPrice),
				fmt.Sprintf("%.2f", p.Price),
			}, widths, aligns)
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=price-report.pdf")
		doc.Output(ctx.Writer)
	}
}
