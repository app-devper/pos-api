package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetTaxInvoicePDF(orderEntity repositories.IOrder, settingEntity repositories.ISetting) gin.HandlerFunc {
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
		companyTaxId := ""
		showCredit := true
		if setting != nil {
			if setting.CompanyName != "" {
				companyName = setting.CompanyName
			}
			companyAddress = setting.CompanyAddress
			companyPhone = setting.CompanyPhone
			companyTaxId = setting.CompanyTaxId
			showCredit = setting.ShowCredit
		}

		doc := pdf.NewPDF()
		doc.AddPage()

		pdf.AddHeader(doc, companyName, companyAddress, companyPhone, "Tax Invoice / Receipt")

		doc.SetFont("Arial", "", 9)
		if companyTaxId != "" {
			doc.CellFormat(0, 5, fmt.Sprintf("Tax ID: %s", companyTaxId), "", 1, "C", false, 0, "")
		}
		doc.Ln(2)

		doc.CellFormat(95, 5, fmt.Sprintf("Invoice No: %s", order.Code), "", 0, "L", false, 0, "")
		doc.CellFormat(95, 5, fmt.Sprintf("Date: %s", order.CreatedDate.Format("02/01/2006")), "", 1, "R", false, 0, "")
		if order.CustomerName != "" {
			doc.CellFormat(0, 5, fmt.Sprintf("Customer: %s", order.CustomerName), "", 1, "L", false, 0, "")
		}
		doc.Ln(3)

		headers := []string{"#", "Description", "Qty", "Unit Price", "Amount"}
		widths := []float64{10, 80, 20, 40, 40}
		aligns := []string{"C", "L", "C", "R", "R"}
		pdf.AddTableHeader(doc, headers, widths)

		subtotal := 0.0
		for i, item := range order.Items {
			amount := item.Price - item.Discount
			subtotal += amount
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				item.Product.Name,
				fmt.Sprintf("%d", item.Quantity),
				fmt.Sprintf("%.2f", item.Price/float64(item.Quantity)),
				fmt.Sprintf("%.2f", amount),
			}, widths, aligns)
		}

		doc.Ln(3)
		totalWidth := float64(190)

		pdf.AddSummaryLine(doc, "Subtotal:", fmt.Sprintf("%.2f", subtotal), totalWidth)
		if order.Discount > 0 {
			pdf.AddSummaryLine(doc, "Discount:", fmt.Sprintf("-%.2f", order.Discount), totalWidth)
		}

		vatBase := order.Total / 1.07
		vat := order.Total - vatBase
		pdf.AddSummaryLine(doc, "Before VAT:", fmt.Sprintf("%.2f", vatBase), totalWidth)
		pdf.AddSummaryLine(doc, "VAT 7%:", fmt.Sprintf("%.2f", vat), totalWidth)
		pdf.AddSummaryLine(doc, "Grand Total:", fmt.Sprintf("%.2f", order.Total), totalWidth)

		creditText := ""
		if showCredit {
			creditText = "Powered by POS System"
		}
		pdf.AddFooter(doc, creditText, false)

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=tax-invoice-%s.pdf", order.Code))
		err = doc.Output(ctx.Writer)
		if err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
