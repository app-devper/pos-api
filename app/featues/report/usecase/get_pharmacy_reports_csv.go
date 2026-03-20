package usecase

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

// helper to resolve order map from order range
func buildOrderMap(orderEntity repositories.IOrder, orderRange request.GetOrderRange) (map[string]*entities.Order, error) {
	orders, err := orderEntity.GetOrderRange(orderRange)
	if err != nil {
		return nil, err
	}
	om := make(map[string]*entities.Order, len(orders))
	for i := range orders {
		om[orders[i].Id.Hex()] = &orders[i]
	}
	return om, nil
}

func GetKHY9CSV(receiveEntity repositories.IReceive, productEntity repositories.IProduct, supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")

		receiveRange := request.GetReceiveRange{
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
			BranchId:  branchId,
		}
		receives, err := receiveEntity.GetReceives(receiveRange)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}

		productIdSet := make(map[string]struct{})
		supplierIdSet := make(map[string]struct{})
		for _, recv := range receives {
			if !recv.SupplierId.IsZero() {
				supplierIdSet[recv.SupplierId.Hex()] = struct{}{}
			}
			for _, item := range recv.Items {
				productIdSet[item.ProductId.Hex()] = struct{}{}
			}
		}
		productIds := make([]string, 0, len(productIdSet))
		for id := range productIdSet {
			productIds = append(productIds, id)
		}
		productList, _ := productEntity.GetProductsByIds(productIds)
		productMap := make(map[string]*entities.Product, len(productList))
		for i := range productList {
			productMap[productList[i].Id.Hex()] = &productList[i]
		}

		supplierMap := make(map[string]string)
		for sid := range supplierIdSet {
			if sup, err := supplierEntity.GetSupplierById(sid); err == nil && sup != nil {
				supplierMap[sid] = sup.Name
			}
		}

		ctx.Header("Content-Type", "text/csv; charset=utf-8")
		ctx.Header("Content-Disposition", "attachment; filename=khy9-report.csv")
		ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

		w := csv.NewWriter(ctx.Writer)
		w.Write([]string{"#", "วันที่", "เลขที่", "ชื่อยา", "ชื่อสามัญ", "ความแรง", "ล็อต", "วันหมดอายุ", "จำนวน", "หน่วย", "ต้นทุน", "ผู้จำหน่าย"})

		row := 1
		for _, recv := range receives {
			supName := supplierMap[recv.SupplierId.Hex()]
			for _, item := range recv.Items {
				product, ok := productMap[item.ProductId.Hex()]
				if !ok || !containsReg(product.DrugRegistrations, "KHY9") {
					continue
				}
				expStr := ""
				if !item.ExpireDate.IsZero() {
					expStr = item.ExpireDate.Format("02/01/2006")
				}
				w.Write([]string{
					fmt.Sprintf("%d", row),
					recv.CreatedDate.Format("02/01/2006"),
					recv.Code,
					product.Name,
					getDrugField(*product, func(d *entities.DrugInfo) string { return d.GenericName }),
					getDrugField(*product, func(d *entities.DrugInfo) string { return d.Strength }),
					item.LotNumber,
					expStr,
					fmt.Sprintf("%d", item.Quantity),
					product.Unit,
					fmt.Sprintf("%.2f", item.CostPrice),
					supName,
				})
				row++
			}
		}
		w.Flush()
	}
}

func generateSalesCSV(ctx *gin.Context, orderEntity repositories.IOrder, branchId string, req pharmacyReportRange, khyKey string, filename string) {
	orderRange := request.GetOrderRange{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		BranchId:  branchId,
	}
	orderMap, err := buildOrderMap(orderEntity, orderRange)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
		return
	}
	orderItems, err := orderEntity.GetOrderItemRange(orderRange)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
		return
	}

	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", filename))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(ctx.Writer)
	w.Write([]string{"#", "วันที่", "ชื่อยา", "ชื่อสามัญ", "ความแรง", "รูปแบบยา", "จำนวน", "หน่วย", "วิธีใช้", "เภสัชกร", "เลขใบอนุญาต"})

	row := 1
	for _, oi := range orderItems {
		product := oi.Product
		if !containsReg(product.DrugRegistrations, khyKey) {
			continue
		}
		order := orderMap[oi.OrderId.Hex()]
		if order == nil || order.Status != constant.ACTIVE {
			continue
		}
		w.Write([]string{
			fmt.Sprintf("%d", row),
			oi.CreatedDate.Format("02/01/2006"),
			product.Name,
			getDrugField(product, func(d *entities.DrugInfo) string { return d.GenericName }),
			getDrugField(product, func(d *entities.DrugInfo) string { return d.Strength }),
			getDrugField(product, func(d *entities.DrugInfo) string { return d.DosageForm }),
			fmt.Sprintf("%d", oi.Quantity),
			product.Unit,
			getDrugField(product, func(d *entities.DrugInfo) string { return d.Dosage }),
			order.PharmacistName,
			order.LicenseNo,
		})
		row++
	}
	w.Flush()
}

func GetKHY10CSV(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateSalesCSV(ctx, orderEntity, branchId, req, "KHY10", "khy10-report")
	}
}

func GetKHY11CSV(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateSalesCSV(ctx, orderEntity, branchId, req, "KHY11", "khy11-report")
	}
}

func GetKHY12CSV(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateSalesCSV(ctx, orderEntity, branchId, req, "KHY12", "khy12-report")
	}
}

func GetKHY13CSV(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		generateSalesCSV(ctx, orderEntity, branchId, req, "KHY13", "khy13-report")
	}
}
