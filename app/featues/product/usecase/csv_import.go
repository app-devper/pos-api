package usecase

import (
	"encoding/csv"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CSVImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

func ImportCSV(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file, _, err := ctx.Request.FormFile("file")
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, "ไม่พบไฟล์ CSV")
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, "อ่านไฟล์ CSV ไม่สำเร็จ: "+err.Error())
			return
		}

		if len(records) < 2 {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, "ไฟล์ CSV ต้องมีอย่างน้อย 1 แถวข้อมูล (ไม่นับ header)")
			return
		}

		userId := utils.GetUserId(ctx)
		result := CSVImportResult{Total: len(records) - 1}
		var errs []string

		// Expected CSV columns: name, nameEn, serialNumber, unit, price, costPrice, category, description, drugType, drugRegistrations
		for i, row := range records[1:] {
			rowNum := i + 2
			if len(row) < 6 {
				errs = append(errs, "แถว "+strconv.Itoa(rowNum)+": ข้อมูลไม่ครบ (ต้องมีอย่างน้อย 6 คอลัมน์)")
				result.Failed++
				continue
			}

			name := strings.TrimSpace(row[0])
			nameEn := ""
			if len(row) > 1 {
				nameEn = strings.TrimSpace(row[1])
			}
			serialNumber := strings.TrimSpace(row[2])
			unit := strings.TrimSpace(row[3])
			price, _ := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
			costPrice, _ := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
			category := ""
			if len(row) > 6 {
				category = strings.TrimSpace(row[6])
			}
			description := ""
			if len(row) > 7 {
				description = strings.TrimSpace(row[7])
			}
			drugType := ""
			if len(row) > 8 {
				drugType = strings.TrimSpace(row[8])
			}
			var drugRegistrations []string
			if len(row) > 9 && strings.TrimSpace(row[9]) != "" {
				for _, dr := range strings.Split(row[9], "|") {
					drugRegistrations = append(drugRegistrations, strings.TrimSpace(dr))
				}
			}

			if name == "" || serialNumber == "" {
				errs = append(errs, "แถว "+strconv.Itoa(rowNum)+": ชื่อสินค้าหรือรหัสสินค้าว่าง")
				result.Failed++
				continue
			}

			// Skip duplicate serial numbers
			if existing, _ := productEntity.GetProductBySerialNumber(serialNumber); existing != nil {
				errs = append(errs, "แถว "+strconv.Itoa(rowNum)+": รหัสสินค้า "+serialNumber+" มีอยู่แล้วในระบบ")
				result.Failed++
				continue
			}

			var drugInfo *request.RequestDrugInfo
			if drugType != "" {
				drugInfo = &request.RequestDrugInfo{DrugType: drugType}
			}

			_, err := productEntity.CreateProduct(request.Product{
				Name:              name,
				NameEn:            nameEn,
				SerialNumber:      serialNumber,
				Unit:              unit,
				Price:             price,
				CostPrice:         costPrice,
				Category:          category,
				Description:       description,
				Status:            "ACTIVE",
				DrugInfo:          drugInfo,
				DrugRegistrations: drugRegistrations,
				CreatedBy:         userId,
			})
			if err != nil {
				errs = append(errs, "แถว "+strconv.Itoa(rowNum)+": "+err.Error())
				result.Failed++
				continue
			}
			result.Success++
		}

		result.Errors = errs
		ctx.JSON(http.StatusOK, result)
	}
}
