package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

type barcodeRequest struct {
	ProductIds []string `json:"productIds" binding:"required"`
	Copies     int      `json:"copies"`
}

func GetBarcodePDF(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := barcodeRequest{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		if req.Copies <= 0 {
			req.Copies = 1
		}

		type labelData struct {
			Name         string
			SerialNumber string
			Price        float64
			Unit         string
		}

		var labels []labelData
		for _, pid := range req.ProductIds {
			product, err := productEntity.GetProductById(pid)
			if err != nil {
				continue
			}
			for i := 0; i < req.Copies; i++ {
				labels = append(labels, labelData{
					Name:         product.Name,
					SerialNumber: product.SerialNumber,
					Price:        product.Price,
					Unit:         product.Unit,
				})
			}
		}

		labelW := 50.0
		labelH := 25.0
		cols := 3
		rows := 10
		marginX := 12.0
		marginY := 10.0
		gapX := 6.0
		gapY := 2.7

		doc := fpdf.New("P", "mm", "A4", "")
		doc.SetAutoPageBreak(false, 0)

		idx := 0
		for idx < len(labels) {
			doc.AddPage()
			for r := 0; r < rows && idx < len(labels); r++ {
				for c := 0; c < cols && idx < len(labels); c++ {
					label := labels[idx]
					x := marginX + float64(c)*(labelW+gapX)
					y := marginY + float64(r)*(labelH+gapY)

					doc.Rect(x, y, labelW, labelH, "D")

					doc.SetFont("Arial", "B", 7)
					doc.SetXY(x+1, y+1)
					nameStr := label.Name
					if len(nameStr) > 25 {
						nameStr = nameStr[:25] + ".."
					}
					doc.CellFormat(labelW-2, 4, nameStr, "", 1, "C", false, 0, "")

					doc.SetFont("Arial", "", 6)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 3, fmt.Sprintf("Unit: %s", label.Unit), "", 1, "C", false, 0, "")

					doc.SetFont("Courier", "B", 10)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 6, label.SerialNumber, "", 1, "C", false, 0, "")

					barcodeY := doc.GetY()
					drawCode128(doc, x+3, barcodeY, labelW-6, 5, label.SerialNumber)

					doc.SetFont("Arial", "B", 9)
					doc.SetXY(x+1, y+labelH-5)
					doc.CellFormat(labelW-2, 4, fmt.Sprintf("%.2f", label.Price), "", 1, "R", false, 0, "")

					idx++
				}
			}
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=barcode-labels.pdf")
		if err := doc.Output(ctx.Writer); err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
		}
	}
}

func GetPriceTagPDF(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProduct{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			req = request.GetProduct{}
		}

		products, err := productEntity.GetProductAll(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		labelW := 60.0
		labelH := 30.0
		cols := 3
		rows := 9
		marginX := 7.0
		marginY := 7.0
		gapX := 3.0
		gapY := 1.0

		doc := fpdf.New("P", "mm", "A4", "")
		doc.SetAutoPageBreak(false, 0)

		idx := 0
		for idx < len(products) {
			doc.AddPage()
			for r := 0; r < rows && idx < len(products); r++ {
				for c := 0; c < cols && idx < len(products); c++ {
					p := products[idx]
					x := marginX + float64(c)*(labelW+gapX)
					y := marginY + float64(r)*(labelH+gapY)

					doc.Rect(x, y, labelW, labelH, "D")

					doc.SetFont("Arial", "B", 7)
					doc.SetXY(x+1, y+1)
					nameStr := p.Name
					if len(nameStr) > 30 {
						nameStr = nameStr[:30] + ".."
					}
					doc.CellFormat(labelW-2, 4, nameStr, "", 1, "C", false, 0, "")

					doc.SetFont("Arial", "", 6)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 3, fmt.Sprintf("SN: %s | %s", p.SerialNumber, p.Unit), "", 1, "C", false, 0, "")

					doc.SetFont("Courier", "B", 8)
					doc.SetX(x + 1)
					doc.CellFormat(labelW-2, 5, p.SerialNumber, "", 1, "C", false, 0, "")

					barcodeY := doc.GetY()
					drawCode128(doc, x+3, barcodeY, labelW-6, 5, p.SerialNumber)

					doc.SetFont("Arial", "B", 12)
					doc.SetXY(x+1, y+labelH-7)
					doc.CellFormat(labelW-2, 6, fmt.Sprintf("%.2f", p.Price), "", 1, "C", false, 0, "")

					idx++
				}
			}
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=price-tags.pdf")
		if err := doc.Output(ctx.Writer); err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
		}
	}
}

func drawCode128(pdf *fpdf.Fpdf, x, y, w, h float64, code string) {
	if code == "" {
		return
	}
	barWidth := w / float64(len(code)*11+35)
	if barWidth < 0.2 {
		barWidth = 0.2
	}
	cx := x
	for i, ch := range code {
		if i%2 == 0 {
			pdf.SetFillColor(0, 0, 0)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		bw := barWidth * float64(1+(int(ch)%3))
		if cx+bw > x+w {
			break
		}
		pdf.Rect(cx, y, bw, h, "F")
		cx += bw
	}
}
