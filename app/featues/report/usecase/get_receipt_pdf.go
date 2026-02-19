package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetReceiptPDF(orderEntity repositories.IOrder, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		branchId := ctx.GetString("BranchId")

		order, err := orderEntity.GetOrderDetailById(orderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "POS System"
		companyAddress := ""
		companyPhone := ""
		footerText := ""
		showCredit := true
		if setting != nil {
			if setting.CompanyName != "" {
				companyName = setting.CompanyName
			}
			companyAddress = setting.CompanyAddress
			companyPhone = setting.CompanyPhone
			footerText = setting.ReceiptFooter
			showCredit = setting.ShowCredit
		}

		doc := pdf.NewPDF()
		doc.AddPage()

		pdf.AddHeader(doc, companyName, companyAddress, companyPhone, "Receipt / Invoice")

		// Order info
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(95, 5, fmt.Sprintf("Order: %s", order.Code), "", 0, "L", false, 0, "")
		doc.CellFormat(95, 5, fmt.Sprintf("Date: %s", order.CreatedDate.Format("02/01/2006 15:04")), "", 1, "R", false, 0, "")
		if order.CustomerCode != "" {
			doc.CellFormat(0, 5, fmt.Sprintf("Customer: %s (%s)", order.CustomerName, order.CustomerCode), "", 1, "L", false, 0, "")
		}
		doc.Ln(3)

		// Items table
		headers := []string{"#", "Item", "Qty", "Price", "Discount", "Total"}
		widths := []float64{10, 70, 20, 30, 30, 30}
		aligns := []string{"C", "L", "C", "R", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, item := range order.Items {
			itemTotal := item.Price - item.Discount
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				item.Product.Name,
				fmt.Sprintf("%d", item.Quantity),
				fmt.Sprintf("%.2f", item.Price),
				fmt.Sprintf("%.2f", item.Discount),
				fmt.Sprintf("%.2f", itemTotal),
			}, widths, aligns)
		}

		doc.Ln(3)
		totalWidth := float64(190)

		if order.Discount > 0 {
			pdf.AddSummaryLine(doc, "Subtotal:", fmt.Sprintf("%.2f", order.Total+order.Discount), totalWidth)
			pdf.AddSummaryLine(doc, "Discount:", fmt.Sprintf("-%.2f", order.Discount), totalWidth)
		}
		pdf.AddSummaryLine(doc, "Total:", fmt.Sprintf("%.2f", order.Total), totalWidth)
		pdf.AddSummaryLine(doc, "Paid:", fmt.Sprintf("%.2f", order.Payment.Amount), totalWidth)
		pdf.AddSummaryLine(doc, "Change:", fmt.Sprintf("%.2f", order.Payment.Change), totalWidth)

		pdf.AddFooter(doc, footerText, showCredit)

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=receipt-%s.pdf", order.Code))
		err = doc.Output(ctx.Writer)
		if err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
