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

func GetReceiveSummaryPDF(receiveEntity repositories.IReceive, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetReceiveRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		req.BranchId = branchId

		receives, err := receiveEntity.GetReceives(req)
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
		pdf.AddHeader(doc, companyName, "", "", "Receive Summary Report")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Period: %s - %s", req.StartDate.Format("02/01/2006"), req.EndDate.Format("02/01/2006")), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Date", "Code", "Reference", "Items", "Total Cost"}
		widths := []float64{10, 30, 30, 40, 20, 40}
		aligns := []string{"C", "L", "L", "L", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		totalCost := 0.0
		totalItems := 0
		for i, recv := range receives {
			itemCount := len(recv.Items)
			totalItems += itemCount
			totalCost += recv.TotalCost
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				recv.CreatedDate.Format("02/01/2006"),
				recv.Code,
				recv.Reference,
				fmt.Sprintf("%d", itemCount),
				fmt.Sprintf("%.2f", recv.TotalCost),
			}, widths, aligns)
		}

		doc.Ln(3)
		tw := float64(170)
		pdf.AddSummaryLine(doc, "Total Receives:", fmt.Sprintf("%d", len(receives)), tw)
		pdf.AddSummaryLine(doc, "Total Items:", fmt.Sprintf("%d", totalItems), tw)
		pdf.AddSummaryLine(doc, "Total Cost:", fmt.Sprintf("%.2f", totalCost), tw)

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=receive-summary.pdf")
		doc.Output(ctx.Writer)
	}
}
