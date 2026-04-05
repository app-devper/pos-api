package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  branchId,
				"startDate": req.StartDate,
				"endDate":   req.EndDate,
			}).Error("get sales report data failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		totalOrders := 0
		totalRevenue := 0.0
		totalCost := 0.0

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

		rowIndex := 0
		for _, order := range orders {
			if order.Status != constant.ACTIVE {
				continue
			}
			totalOrders++
			totalRevenue += order.Total
			totalCost += order.TotalCost
			i := rowIndex
			rowIndex++
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

		summaryRow := rowIndex + 3
		if totalOrders > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow), "Total Orders:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow), totalOrders)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+1), "Total Revenue:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+1), totalRevenue)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+2), "Total Cost:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+2), totalCost)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow+3), "Total Profit:")
			f.SetCellValue(sheet, fmt.Sprintf("G%d", summaryRow+3), totalRevenue-totalCost)
		}

		for i := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheet, col, col, 18)
		}

		ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sales-report-%s.xlsx",
			req.StartDate.Format("20060102")))
		if err := f.Write(ctx.Writer); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  branchId,
				"startDate": req.StartDate,
				"endDate":   req.EndDate,
			}).Error("write sales report excel failed")
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}
