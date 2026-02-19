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

func GetSalesReportPDF(orderEntity repositories.IOrder, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetOrderRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		req.BranchId = branchId

		orders, err := orderEntity.GetOrderRange(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		summary, _ := orderEntity.GetOrderSummary(req)

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "POS System"
		companyAddress := ""
		companyPhone := ""
		showCredit := true
		if setting != nil {
			if setting.CompanyName != "" {
				companyName = setting.CompanyName
			}
			companyAddress = setting.CompanyAddress
			companyPhone = setting.CompanyPhone
			showCredit = setting.ShowCredit
		}

		doc := pdf.NewPDF()
		doc.AddPage()

		dateRange := fmt.Sprintf("%s - %s",
			req.StartDate.Format("02/01/2006"),
			req.EndDate.Format("02/01/2006"))
		pdf.AddHeader(doc, companyName, companyAddress, companyPhone, "Sales Report")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Period: %s", dateRange), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Code", "Date", "Customer", "Type", "Total"}
		widths := []float64{10, 30, 35, 50, 25, 40}
		aligns := []string{"C", "L", "L", "L", "C", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, order := range orders {
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				order.Code,
				order.CreatedDate.Format("02/01/2006 15:04"),
				order.CustomerName,
				order.Type,
				fmt.Sprintf("%.2f", order.Total),
			}, widths, aligns)
		}

		doc.Ln(5)
		totalWidth := float64(190)
		if summary != nil {
			pdf.AddSummaryLine(doc, "Total Orders:", fmt.Sprintf("%d", summary.TotalOrders), totalWidth)
			pdf.AddSummaryLine(doc, "Total Revenue:", fmt.Sprintf("%.2f", summary.TotalRevenue), totalWidth)
			pdf.AddSummaryLine(doc, "Total Cost:", fmt.Sprintf("%.2f", summary.TotalCost), totalWidth)
			pdf.AddSummaryLine(doc, "Total Profit:", fmt.Sprintf("%.2f", summary.TotalProfit), totalWidth)
		}

		creditText := ""
		if showCredit {
			creditText = "Powered by POS System"
		}
		pdf.AddFooter(doc, creditText, false)

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=sales-report-%s.pdf",
			req.StartDate.Format("20060102")))
		err = doc.Output(ctx.Writer)
		if err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
