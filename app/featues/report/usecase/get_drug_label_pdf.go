package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

func GetDrugLabelPDF(dispensingEntity repositories.IDispensingLog, settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logId := ctx.Param("logId")
		branchId := ctx.GetString("BranchId")

		dispLog, err := dispensingEntity.GetDispensingLogById(logId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		setting, _ := settingEntity.GetSettingByBranchId(branchId)
		companyName := "Pharmacy"
		if setting != nil && setting.CompanyName != "" {
			companyName = setting.CompanyName
		}

		labelW := 70.0
		labelH := 35.0
		cols := 2
		rows := 8
		marginX := 10.0
		marginY := 10.0
		gapX := 5.0
		gapY := 2.0

		doc := fpdf.New("P", "mm", "A4", "")
		doc.SetAutoPageBreak(false, 0)
		doc.AddPage()

		itemIdx := 0
		totalItems := len(dispLog.Items)

		for itemIdx < totalItems {
			for r := 0; r < rows && itemIdx < totalItems; r++ {
				for c := 0; c < cols && itemIdx < totalItems; c++ {
					item := dispLog.Items[itemIdx]
					x := marginX + float64(c)*(labelW+gapX)
					y := marginY + float64(r)*(labelH+gapY)

					doc.SetFont("Arial", "B", 7)
					doc.SetXY(x+1, y+1)
					doc.CellFormat(labelW-2, 4, companyName, "", 1, "C", false, 0, "")

					doc.SetFont("Arial", "", 6)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 3, fmt.Sprintf("Pharmacist: %s (Lic: %s)", dispLog.PharmacistName, dispLog.LicenseNo), "", 1, "L", false, 0, "")

					doc.SetFont("Arial", "B", 7)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 4, item.ProductName, "", 1, "L", false, 0, "")

					doc.SetFont("Arial", "", 6)
					if item.GenericName != "" {
						doc.SetX(x + 1)
						doc.CellFormat(labelW-2, 3, fmt.Sprintf("(%s)", item.GenericName), "", 1, "L", false, 0, "")
					}

					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 3, fmt.Sprintf("Qty: %d %s", item.Quantity, item.Unit), "", 1, "L", false, 0, "")

					if item.Dosage != "" {
						doc.SetFont("Arial", "B", 6)
						doc.SetX(x + 1)
						doc.CellFormat(labelW-2, 3, item.Dosage, "", 1, "L", false, 0, "")
					}

					if item.LotNumber != "" {
						doc.SetFont("Arial", "", 5)
						doc.SetX(x + 1)
						doc.CellFormat(labelW-2, 3, fmt.Sprintf("Lot: %s", item.LotNumber), "", 1, "L", false, 0, "")
					}

					doc.SetFont("Arial", "", 5)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 3, fmt.Sprintf("Date: %s", dispLog.CreatedDate.Format("02/01/2006")), "", 1, "L", false, 0, "")

					doc.Rect(x, y, labelW, labelH, "D")
					itemIdx++
				}
			}
			if itemIdx < totalItems {
				doc.AddPage()
			}
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=drug-labels-%s.pdf", logId))
		if err := doc.Output(ctx.Writer); err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
		}
	}
}
