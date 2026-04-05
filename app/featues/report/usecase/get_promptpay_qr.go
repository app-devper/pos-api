package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/pdf"
	"pos/app/data/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/sirupsen/logrus"
)

func GetPromptPayQR(settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		amountStr := ctx.Query("amount")
		amount, err := parsePromptPayAmount(amountStr)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}

		setting, err := settingEntity.GetSettingByBranchId(branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId": branchId,
				"amount":   amount,
			}).Error("get promptpay qr setting failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		companyName := "POS System"
		promptPayId := ""
		if setting != nil {
			if setting.CompanyName != "" {
				companyName = setting.CompanyName
			}
			if setting.PromptPayId != "" {
				promptPayId = setting.PromptPayId
			}
		}

		if promptPayId == "" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, "PromptPay ID not configured in settings")
			return
		}

		payload := generatePromptPayPayload(promptPayId, amount)

		doc := fpdf.New("P", "mm", "A4", "")
		pdf.InitFont(doc)
		doc.AddPage()
		doc.SetFont(pdf.FontFamily, "B", 14)
		doc.CellFormat(0, 10, companyName, "", 1, "C", false, 0, "")
		doc.Ln(3)
		doc.SetFont(pdf.FontFamily, "B", 12)
		doc.CellFormat(0, 8, "PromptPay QR Code", "", 1, "C", false, 0, "")
		doc.Ln(2)
		doc.SetFont(pdf.FontFamily, "", 10)
		doc.CellFormat(0, 6, fmt.Sprintf("PromptPay ID: %s", promptPayId), "", 1, "C", false, 0, "")
		if amount > 0 {
			doc.CellFormat(0, 6, fmt.Sprintf("Amount: %.2f THB", amount), "", 1, "C", false, 0, "")
		}
		doc.Ln(5)

		doc.SetFont(pdf.FontFamily, "", 8)
		doc.CellFormat(0, 5, "EMVCo Payload:", "", 1, "C", false, 0, "")
		doc.CellFormat(0, 5, payload, "", 1, "C", false, 0, "")
		doc.Ln(5)
		doc.SetFont(pdf.FontFamily, "I", 9)
		doc.CellFormat(0, 5, "Use this payload with a QR generator to create scannable QR code", "", 1, "C", false, 0, "")

		ctx.Header("Content-Type", "application/pdf")
		ctx.Header("Content-Disposition", "inline; filename=promptpay-qr.pdf")
		if err := doc.Output(ctx.Writer); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId": branchId,
				"amount":   amount,
			}).Error("write promptpay qr pdf failed")
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.RP_INTERNAL_001, err.Error())
			return
		}
	}
}

func GetPromptPayPayload(settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		amountStr := ctx.Query("amount")
		amount, err := parsePromptPayAmount(amountStr)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}

		setting, err := settingEntity.GetSettingByBranchId(branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId": branchId,
				"amount":   amount,
			}).Error("get promptpay payload setting failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		promptPayId := ""
		if setting != nil && setting.PromptPayId != "" {
			promptPayId = setting.PromptPayId
		}

		if promptPayId == "" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, "PromptPay ID not configured in settings")
			return
		}

		payload := generatePromptPayPayload(promptPayId, amount)
		ctx.JSON(http.StatusOK, gin.H{"payload": payload, "promptPayId": promptPayId, "amount": amount})
	}
}

func parsePromptPayAmount(amountStr string) (float64, error) {
	if amountStr == "" {
		return 0, nil
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("amount is not valid")
	}
	return amount, nil
}

func generatePromptPayPayload(promptPayId string, amount float64) string {
	formatIndicator := "000201"
	pointOfInit := "010212"

	idType := "01"
	if len(promptPayId) == 13 {
		idType = "02"
	}
	merchantId := fmt.Sprintf("%s%02d%s", idType, len(promptPayId), promptPayId)
	aid := "0016A000000677010111"
	merchantInfo := fmt.Sprintf("%s%s", aid, merchantId)
	tag29 := fmt.Sprintf("29%02d%s", len(merchantInfo), merchantInfo)

	countryCode := "5802TH"
	currency := "5303764"

	amountTag := ""
	if amount > 0 {
		amountStr := fmt.Sprintf("%.2f", amount)
		amountTag = fmt.Sprintf("54%02d%s", len(amountStr), amountStr)
	}

	payload := formatIndicator + pointOfInit + tag29 + countryCode + currency + amountTag
	crc := crc16(payload + "6304")
	payload = payload + fmt.Sprintf("6304%04X", crc)
	return payload
}

func crc16(data string) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
