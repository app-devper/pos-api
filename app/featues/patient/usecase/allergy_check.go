package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"strings"

	"github.com/gin-gonic/gin"
)

func AllergyCheck(patientEntity repositories.IPatient, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		patientId := ctx.Param("id")
		req := request.AllergyCheck{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_001, err.Error())
			return
		}

		patient, err := patientEntity.GetPatientById(patientId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, "patient not found")
			return
		}

		if len(patient.Allergies) == 0 {
			ctx.JSON(http.StatusOK, []request.AllergyCheckResult{})
			return
		}

		allergyMap := make(map[string]int)
		for i, a := range patient.Allergies {
			allergyMap[strings.ToLower(a.DrugName)] = i
		}

		products, _ := productEntity.GetProductsByIds(req.ProductIds)

		var warnings []request.AllergyCheckResult
		for _, p := range products {
			if p.DrugInfo == nil {
				continue
			}
			pid := p.Id.Hex()
			namesToCheck := []string{
				strings.ToLower(p.Name),
				strings.ToLower(p.DrugInfo.GenericName),
			}
			for _, name := range namesToCheck {
				if name == "" {
					continue
				}
				if idx, ok := allergyMap[name]; ok {
					warnings = append(warnings, request.AllergyCheckResult{
						ProductId:   pid,
						ProductName: p.Name,
						DrugName:    patient.Allergies[idx].DrugName,
						Reaction:    patient.Allergies[idx].Reaction,
						Severity:    patient.Allergies[idx].Severity,
					})
					break
				}
			}
		}

		if warnings == nil {
			warnings = []request.AllergyCheckResult{}
		}
		ctx.JSON(http.StatusOK, warnings)
	}
}
