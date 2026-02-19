package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetCustomerHistoryPDF(customerHistoryEntity repositories.ICustomerHistory, customerEntity repositories.ICustomer, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerCode := ctx.Param("customerCode")
		branchId := ctx.GetString("BranchId")

		customer, err := customerEntity.GetCustomerByCode(customerCode)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		histories, err := customerHistoryEntity.GetCustomerHistories(customerCode, branchId)
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
		pdf.AddHeader(doc, companyName, "", "", "Customer History Report")
		doc.SetFont("Arial", "", 9)
		doc.CellFormat(0, 5, fmt.Sprintf("Customer: %s (%s)", customer.Name, customer.Code), "", 1, "C", false, 0, "")
		doc.Ln(3)

		headers := []string{"#", "Date", "Type", "Description", "Reference"}
		widths := []float64{10, 35, 35, 60, 50}
		aligns := []string{"C", "L", "L", "L", "L"}
		pdf.AddTableHeader(doc, headers, widths)

		for i, h := range histories {
			pdf.AddTableRow(doc, []string{
				fmt.Sprintf("%d", i+1),
				h.CreatedDate.Format("02/01/2006 15:04"),
				h.Type,
				h.Description,
				h.Reference,
			}, widths, aligns)
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=customer-history-%s.pdf", customerCode))
		doc.Output(ctx.Writer)
	}
}
