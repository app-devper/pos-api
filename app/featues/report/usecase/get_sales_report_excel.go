package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetSalesReportExcel(orderEntity repositories.IOrder) gin.HandlerFunc {
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

		f := excelize.NewFile()
		sheet := "Sales Report"
		f.SetSheetName("Sheet1", sheet)

		headers := []string{"#", "Code", "Date", "Customer Code", "Customer Name", "Type", "Total", "Total Cost", "Discount"}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", string(rune('A'+len(headers)-1))), style)

		for i, order := range orders {
			row := i + 2
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), order.Code)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), order.CreatedDate.Format("02/01/2006 15:04"))
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), order.CustomerCode)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), order.CustomerName)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), order.Type)
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), order.Total)
			f.SetCellValue(sheet, fmt.Sprintf("H%d", row), order.TotalCost)
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), order.Discount)
		}

		if summary != nil {
			summaryRow := len(orders) + 3
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow), "Total Orders:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow), summary.TotalOrders)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+1), "Total Revenue:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+1), summary.TotalRevenue)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+2), "Total Cost:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+2), summary.TotalCost)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+3), "Total Profit:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+3), summary.TotalProfit)
		}

		for i := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheet, col, col, 18)
		}

		ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sales-report-%s.xlsx",
			req.StartDate.Format("20060102")))
		if err := f.Write(ctx.Writer); err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
