package pdf

import (
	"github.com/go-pdf/fpdf"
)

const (
	FontFamily = "Arial"
	FontSize   = 10
	HeaderSize = 14
	TitleSize  = 12
)

func NewPDF() *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetFont(FontFamily, "", FontSize)
	return pdf
}

func AddHeader(pdf *fpdf.Fpdf, companyName string, companyAddress string, companyPhone string, title string) {
	pdf.SetFont(FontFamily, "B", HeaderSize)
	pdf.CellFormat(0, 8, companyName, "", 1, "C", false, 0, "")
	pdf.SetFont(FontFamily, "", FontSize)
	if companyAddress != "" {
		pdf.CellFormat(0, 5, companyAddress, "", 1, "C", false, 0, "")
	}
	if companyPhone != "" {
		pdf.CellFormat(0, 5, "Tel: "+companyPhone, "", 1, "C", false, 0, "")
	}
	pdf.Ln(3)
	pdf.SetFont(FontFamily, "B", TitleSize)
	pdf.CellFormat(0, 8, title, "", 1, "C", false, 0, "")
	pdf.Ln(3)
}

func AddFooter(pdf *fpdf.Fpdf, footerText string, showCredit bool) {
	pdf.SetY(-25)
	pdf.SetFont(FontFamily, "", 8)
	if footerText != "" {
		pdf.CellFormat(0, 5, footerText, "", 1, "C", false, 0, "")
	}
	if showCredit {
		pdf.CellFormat(0, 5, "Powered by POS System", "", 1, "C", false, 0, "")
	}
}

func AddTableHeader(pdf *fpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont(FontFamily, "B", FontSize)
	pdf.SetFillColor(220, 220, 220)
	for i, header := range headers {
		pdf.CellFormat(widths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont(FontFamily, "", FontSize)
}

func AddTableRow(pdf *fpdf.Fpdf, cells []string, widths []float64, aligns []string) {
	for i, cell := range cells {
		align := "L"
		if i < len(aligns) {
			align = aligns[i]
		}
		pdf.CellFormat(widths[i], 6, cell, "1", 0, align, false, 0, "")
	}
	pdf.Ln(-1)
}

func AddSummaryLine(pdf *fpdf.Fpdf, label string, value string, totalWidth float64) {
	labelWidth := totalWidth * 0.7
	valueWidth := totalWidth * 0.3
	pdf.CellFormat(labelWidth, 7, label, "", 0, "R", false, 0, "")
	pdf.SetFont(FontFamily, "B", FontSize)
	pdf.CellFormat(valueWidth, 7, value, "", 1, "R", false, 0, "")
	pdf.SetFont(FontFamily, "", FontSize)
}
