package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func resolveUnitName(unitMap map[string]string, unitID string, fallback string) string {
	if unitID == "" {
		return fallback
	}
	if unit, ok := unitMap[unitID]; ok && unit != "" {
		return unit
	}
	return fallback
}

func buildUnitNameMap(productEntity repositories.IProduct, unitIds []string) (map[string]string, error) {
	if len(unitIds) == 0 {
		return map[string]string{}, nil
	}
	units, err := productEntity.GetProductUnitsByIds(unitIds)
	if err != nil {
		return nil, err
	}
	unitMap := make(map[string]string, len(units))
	for _, unit := range units {
		unitMap[unit.Id.Hex()] = unit.Unit
	}
	return unitMap, nil
}

func getReceiveItemsMap(receiveEntity repositories.IReceive, receives []entities.Receive) (map[string][]entities.ReceiveItem, error) {
	if len(receives) == 0 {
		return map[string][]entities.ReceiveItem{}, nil
	}
	receiveIds := make([]string, 0, len(receives))
	for _, recv := range receives {
		receiveIds = append(receiveIds, recv.Id.Hex())
	}
	receiveItems, err := receiveEntity.GetReceiveItemsByReceiveIds(receiveIds)
	if err != nil {
		return nil, err
	}
	receiveItemsMap := make(map[string][]entities.ReceiveItem, len(receives))
	for _, item := range receiveItems {
		receiveItemsMap[item.ReceiveId.Hex()] = append(receiveItemsMap[item.ReceiveId.Hex()], item)
	}
	return receiveItemsMap, nil
}

func getSupplierNameMap(supplierEntity repositories.ISupplier, supplierIds []string) (map[string]string, error) {
	if len(supplierIds) == 0 {
		return map[string]string{}, nil
	}
	suppliers, err := supplierEntity.GetSuppliersByIds(supplierIds)
	if err != nil {
		return nil, err
	}
	supplierMap := make(map[string]string, len(suppliers))
	for _, supplier := range suppliers {
		supplierMap[supplier.Id.Hex()] = supplier.Name
	}
	return supplierMap, nil
}

type pharmacyReportRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

func containsReg(regs []string, key string) bool {
	for _, r := range regs {
		if r == key {
			return true
		}
	}
	return false
}

type pharmacyReportItem struct {
	Date           time.Time  `json:"date"`
	Code           string     `json:"code,omitempty"`
	ProductName    string     `json:"productName"`
	GenericName    string     `json:"genericName,omitempty"`
	LotNumber      string     `json:"lotNumber,omitempty"`
	Quantity       int        `json:"quantity"`
	Unit           string     `json:"unit,omitempty"`
	CostPrice      float64    `json:"costPrice,omitempty"`
	SupplierName   string     `json:"supplierName,omitempty"`
	ExpireDate     *time.Time `json:"expireDate,omitempty"`
	Strength       string     `json:"strength,omitempty"`
	DosageForm     string     `json:"dosageForm,omitempty"`
	Dosage         string     `json:"dosage,omitempty"`
	PharmacistName string     `json:"pharmacistName,omitempty"`
	LicenseNo      string     `json:"licenseNo,omitempty"`
	DrugType       string     `json:"drugType,omitempty"`
}

type pharmacyReportResponse struct {
	Key       string               `json:"key"`
	Title     string               `json:"title"`
	StartDate time.Time            `json:"startDate"`
	EndDate   time.Time            `json:"endDate"`
	Items     []pharmacyReportItem `json:"items"`
}

func getKHY9Items(receiveEntity repositories.IReceive, productEntity repositories.IProduct, supplierEntity repositories.ISupplier, branchId string, req pharmacyReportRange) ([]pharmacyReportItem, error) {
	receiveRange := request.GetReceiveRange{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		BranchId:  branchId,
	}
	receives, err := receiveEntity.GetReceives(receiveRange)
	if err != nil {
		return nil, err
	}

	receiveItemsMap, err := getReceiveItemsMap(receiveEntity, receives)
	if err != nil {
		return nil, err
	}
	productIdSet := make(map[string]struct{})
	supplierIdSet := make(map[string]struct{})
	unitIdSet := make(map[string]struct{})
	for _, recv := range receives {
		if !recv.SupplierId.IsZero() {
			supplierIdSet[recv.SupplierId.Hex()] = struct{}{}
		}
		for _, item := range receiveItemsMap[recv.Id.Hex()] {
			productIdSet[item.ProductId.Hex()] = struct{}{}
			if item.UnitId != "" {
				unitIdSet[item.UnitId] = struct{}{}
			}
		}
	}
	productIds := make([]string, 0, len(productIdSet))
	for id := range productIdSet {
		productIds = append(productIds, id)
	}
	productList, err := productEntity.GetProductsByIds(productIds)
	if err != nil {
		return nil, err
	}
	productMap := make(map[string]*entities.Product, len(productList))
	for i := range productList {
		productMap[productList[i].Id.Hex()] = &productList[i]
	}
	unitIds := make([]string, 0, len(unitIdSet))
	for id := range unitIdSet {
		unitIds = append(unitIds, id)
	}
	unitMap, err := buildUnitNameMap(productEntity, unitIds)
	if err != nil {
		return nil, err
	}

	supplierIds := make([]string, 0, len(supplierIdSet))
	for sid := range supplierIdSet {
		supplierIds = append(supplierIds, sid)
	}
	supplierMap, err := getSupplierNameMap(supplierEntity, supplierIds)
	if err != nil {
		return nil, err
	}

	items := make([]pharmacyReportItem, 0)
	for _, recv := range receives {
		supName := supplierMap[recv.SupplierId.Hex()]
		for _, item := range receiveItemsMap[recv.Id.Hex()] {
			product, ok := productMap[item.ProductId.Hex()]
			if !ok || !containsReg(product.DrugRegistrations, "KHY9") {
				continue
			}
			var expDate *time.Time
			if !item.ExpireDate.IsZero() {
				expDate = &item.ExpireDate
			}
			items = append(items, pharmacyReportItem{
				Date:         recv.CreatedDate,
				Code:         recv.Code,
				ProductName:  product.Name,
				GenericName:  getDrugField(*product, func(d *entities.DrugInfo) string { return d.GenericName }),
				LotNumber:    item.LotNumber,
				Quantity:     item.Quantity,
				Unit:         resolveUnitName(unitMap, item.UnitId, product.Unit),
				CostPrice:    item.CostPrice,
				SupplierName: supName,
				ExpireDate:   expDate,
				Strength:     getDrugField(*product, func(d *entities.DrugInfo) string { return d.Strength }),
				DosageForm:   getDrugField(*product, func(d *entities.DrugInfo) string { return d.DosageForm }),
				DrugType:     getDrugField(*product, func(d *entities.DrugInfo) string { return d.DrugType }),
			})
		}
	}
	return items, nil
}

func getSalesReportItems(orderEntity repositories.IOrder, productEntity repositories.IProduct, branchId string, req pharmacyReportRange, khyKey string) ([]pharmacyReportItem, error) {
	orderRange := request.GetOrderRange{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		BranchId:  branchId,
	}
	orderMap, err := buildOrderMap(orderEntity, orderRange)
	if err != nil {
		return nil, err
	}

	orderItems, err := orderEntity.GetOrderItemRange(orderRange)
	if err != nil {
		return nil, err
	}
	unitIdSet := make(map[string]struct{})
	for _, oi := range orderItems {
		if !oi.UnitId.IsZero() {
			unitIdSet[oi.UnitId.Hex()] = struct{}{}
		}
	}
	unitIds := make([]string, 0, len(unitIdSet))
	for id := range unitIdSet {
		unitIds = append(unitIds, id)
	}
	unitMap, err := buildUnitNameMap(productEntity, unitIds)
	if err != nil {
		return nil, err
	}

	items := make([]pharmacyReportItem, 0)
	for _, oi := range orderItems {
		product := oi.Product
		if !containsReg(product.DrugRegistrations, khyKey) {
			continue
		}
		order := orderMap[oi.OrderId.Hex()]
		if order == nil || order.Status != constant.ACTIVE {
			continue
		}
		items = append(items, pharmacyReportItem{
			Date:           oi.CreatedDate,
			ProductName:    product.Name,
			GenericName:    getDrugField(product, func(d *entities.DrugInfo) string { return d.GenericName }),
			Quantity:       oi.Quantity,
			Unit:           resolveUnitName(unitMap, oi.UnitId.Hex(), product.Unit),
			Dosage:         getDrugField(product, func(d *entities.DrugInfo) string { return d.Dosage }),
			Strength:       getDrugField(product, func(d *entities.DrugInfo) string { return d.Strength }),
			DosageForm:     getDrugField(product, func(d *entities.DrugInfo) string { return d.DosageForm }),
			PharmacistName: order.PharmacistName,
			LicenseNo:      order.LicenseNo,
			DrugType:       getDrugField(product, func(d *entities.DrugInfo) string { return d.DrugType }),
		})
	}
	return items, nil
}

func GetKHY9Data(receiveEntity repositories.IReceive, productEntity repositories.IProduct, supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		items, err := getKHY9Items(receiveEntity, productEntity, supplierEntity, branchId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, pharmacyReportResponse{Key: "khy9", Title: "ข.ย.9 บัญชีการซื้อยา", StartDate: req.StartDate, EndDate: req.EndDate, Items: items})
	}
}

func getDrugField(p entities.Product, fn func(*entities.DrugInfo) string) string {
	if p.DrugInfo != nil {
		return fn(p.DrugInfo)
	}
	return ""
}

func getSalesReportDataHandler(orderEntity repositories.IOrder, productEntity repositories.IProduct, key string, title string, khyKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := pharmacyReportRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		items, err := getSalesReportItems(orderEntity, productEntity, branchId, req, khyKey)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RP_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, pharmacyReportResponse{Key: key, Title: title, StartDate: req.StartDate, EndDate: req.EndDate, Items: items})
	}
}

func GetKHY10Data(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return getSalesReportDataHandler(orderEntity, productEntity, "khy10", "ข.ย.10 บัญชีการขายยาควบคุมพิเศษ", "KHY10")
}

func GetKHY11Data(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return getSalesReportDataHandler(orderEntity, productEntity, "khy11", "ข.ย.11 บัญชีการขายยาอันตราย", "KHY11")
}

func GetKHY12Data(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return getSalesReportDataHandler(orderEntity, productEntity, "khy12", "ข.ย.12 บัญชีการขายยาตามใบสั่งของผู้ประกอบวิชาชีพ", "KHY12")
}

func GetKHY13Data(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return getSalesReportDataHandler(orderEntity, productEntity, "khy13", "ข.ย.13 รายงานการขายยาตามที่เลขาธิการ อย. กำหนด", "KHY13")
}
