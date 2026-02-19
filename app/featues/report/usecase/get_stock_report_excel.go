package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetStockReportExcel(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")

		stocks, err := productEntity.GetStockReport(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		f := excelize.NewFile()
		sheet := "Stock Report"
		f.SetSheetName("Sheet1", sheet)

		headers := []string{"#", "Serial Number", "Product Name", "Unit", "Total Stock", "Total Cost"}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", string(rune('A'+len(headers)-1))), style)

		for i, stock := range stocks {
			row := i + 2
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), stock.SerialNumber)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), stock.Name)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), stock.Unit)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), stock.TotalStock)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), stock.TotalCost)
		}

		for i := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheet, col, col, 20)
		}

		ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		ctx.Header("Content-Disposition", "attachment; filename=stock-report.xlsx")
		if err := f.Write(ctx.Writer); err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
