package constant

const (
	CustomerTypeGeneral    = "General"
	CustomerTypeWholesaler = "Wholesaler"
	CustomerTypeRegular    = "Regular"
)

const (
	PaymentTypeCash      = "CASH"
	PaymentTypeCredit    = "CREDIT"
	PaymentTypePromptPay = "PROMPTPAY"
	PaymentTypeTransfer  = "TRANSFER"
)

func CustomerTypes() []string {
	return []string{CustomerTypeGeneral, CustomerTypeWholesaler, CustomerTypeRegular}
}

const (
	HistoryTypeAddProduct                 = "AddProduct"
	HistoryTypeUpdateProduct              = "UpdateProduct"
	HistoryTypeAddProductUnit             = "AddProductUnit"
	HistoryTypeUpdateProductUnit          = "UpdateProductUnit"
	HistoryTypeRemoveProductUnit          = "RemoveProductUnit"
	HistoryTypeAddProductPrice            = "AddProductPrice"
	HistoryTypeUpdateProductPrice         = "UpdateProductPrice"
	HistoryTypeRemoveProductPrice         = "RemoveProductPrice"
	HistoryTypeAddProductStock            = "AddProductStock"
	HistoryTypeUpdateProductStock         = "UpdateProductStock"
	HistoryTypeRemoveProductStock         = "RemoveProductStock"
	HistoryTypeUpdateProductStockQuantity = "UpdateProductStockQuantity"
	HistoryTypeAddOrderItemProduct        = "AddOrderItemProduct"
	HistoryTypeRemoveOrderItemProduct     = "RemoveOrderItemProduct"
	HistoryTypeStockAdjustment            = "StockAdjustment"
	HistoryTypeProductReturn              = "ProductReturn"
)

const (
	AdjustmentReasonCount   = "นับสต็อก"
	AdjustmentReasonDamaged = "ยาเสียหาย"
	AdjustmentReasonExpired = "ยาหมดอายุ"
	AdjustmentReasonLost    = "สูญหาย"
	AdjustmentReasonOther   = "อื่นๆ"
)

func AdjustmentReasons() []string {
	return []string{
		AdjustmentReasonCount,
		AdjustmentReasonDamaged,
		AdjustmentReasonExpired,
		AdjustmentReasonLost,
		AdjustmentReasonOther,
	}
}
