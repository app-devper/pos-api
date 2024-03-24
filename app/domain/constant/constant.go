package constant

const (
	CustomerTypeGeneral    = "General"
	CustomerTypeWholesaler = "Wholesaler"
	CustomerTypeRegular    = "Regular"
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
)
