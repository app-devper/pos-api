package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"strings"

	"github.com/gin-gonic/gin"
)

type DrugInteractionCheckRequest struct {
	ProductIds []string `json:"productIds" binding:"required"`
}

type DrugInteractionResult struct {
	ProductAId   string `json:"productAId"`
	ProductAName string `json:"productAName"`
	ProductBId   string `json:"productBId"`
	ProductBName string `json:"productBName"`
	Interaction  string `json:"interaction"`
}

func CheckDrugInteractions(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := DrugInteractionCheckRequest{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}

		if len(req.ProductIds) < 2 {
			ctx.JSON(http.StatusOK, gin.H{"interactions": []DrugInteractionResult{}})
			return
		}

		products, err := productEntity.GetProductsByIds(req.ProductIds)
		if err != nil {
			errcode.Abort(ctx, http.StatusInternalServerError, errcode.PD_INTERNAL_001, err.Error())
			return
		}

		var results []DrugInteractionResult
		seen := make(map[string]bool)

		for i := 0; i < len(products); i++ {
			for j := i + 1; j < len(products); j++ {
				idA := products[i].Id.Hex()
				idB := products[j].Id.Hex()
				nameA := products[i].Name
				nameB := products[j].Name

				pairKey := idA + ":" + idB

				// Check A → B
				if products[i].DrugInfo != nil && len(products[i].DrugInfo.DrugInteractions) > 0 {
					genericB := ""
					if products[j].DrugInfo != nil {
						genericB = products[j].DrugInfo.GenericName
					}
					for _, interaction := range products[i].DrugInfo.DrugInteractions {
						if strings.EqualFold(interaction, nameB) || strings.EqualFold(interaction, genericB) || strings.EqualFold(interaction, products[j].SerialNumber) {
							if !seen[pairKey] {
								seen[pairKey] = true
								results = append(results, DrugInteractionResult{
									ProductAId:   idA,
									ProductAName: nameA,
									ProductBId:   idB,
									ProductBName: nameB,
									Interaction:  interaction,
								})
							}
						}
					}
				}

				// Check B → A (only add if pair not already seen)
				if !seen[pairKey] && products[j].DrugInfo != nil && len(products[j].DrugInfo.DrugInteractions) > 0 {
					genericA := ""
					if products[i].DrugInfo != nil {
						genericA = products[i].DrugInfo.GenericName
					}
					for _, interaction := range products[j].DrugInfo.DrugInteractions {
						if strings.EqualFold(interaction, nameA) || strings.EqualFold(interaction, genericA) || strings.EqualFold(interaction, products[i].SerialNumber) {
							if !seen[pairKey] {
								seen[pairKey] = true
								results = append(results, DrugInteractionResult{
									ProductAId:   idB,
									ProductAName: nameB,
									ProductBId:   idA,
									ProductBName: nameA,
									Interaction:  interaction,
								})
							}
						}
					}
				}
			}
		}

		ctx.JSON(http.StatusOK, gin.H{"interactions": results})
	}
}
