package request

import (
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"strconv"
)

type ProductHistory struct {
	ProductId   string  `json:"productId" binding:"required"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Import      int     `json:"import"`
	Quantity    int     `json:"quantity"`
	CostPrice   float64 `json:"costPrice"`
	Price       float64 `json:"price"`
	Balance     int     `json:"balance"`
	CreatedBy   string  `json:"createdBy"`
}

func AddProductHistory(productId string, product Product) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeAddProduct,
		Description: "เพิ่มสินค้า " + product.Name,
		Unit:        product.Unit,
		CreatedBy:   product.CreatedBy,
	}
}

func UpdateProductHistory(productId string, product UpdateProduct) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeUpdateProduct,
		Description: "แก้ไขสินค้า " + product.Name,
		Unit:        product.Unit,
		CreatedBy:   product.UpdatedBy,
	}
}

func AddProductUnitHistory(productId string, unit ProductUnit) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeAddProductUnit,
		Description: "เพิ่มหน่วยสินค้า " + unit.Unit + " ขนาด " + strconv.Itoa(unit.Size) + "Barcode: " + unit.Barcode,
		Unit:        unit.Unit,
		CreatedBy:   unit.UpdatedBy,
	}
}

func UpdateProductUnitHistory(productId string, unit ProductUnit) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeUpdateProductUnit,
		Description: "แก้ไขหน่วยสินค้า " + unit.Unit,
		Unit:        unit.Unit,
		CreatedBy:   unit.UpdatedBy,
	}
}

func RemoveProductUnitHistory(productId string, unit *entities.ProductUnit, createdBy string) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeRemoveProductUnit,
		Description: "ลบหน่วยสินค้า " + unit.Unit + " ขนาด " + strconv.Itoa(unit.Size) + "Barcode: " + unit.Barcode,
		Unit:        unit.Unit,
		CreatedBy:   createdBy,
	}
}

func AddProductPriceHistory(productId string, unit string, price ProductPrice) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeAddProductPrice,
		Description: "เพิ่มราคาสินค้า " + price.CustomerType,
		Unit:        unit,
		CreatedBy:   price.UpdatedBy,
	}
}

func UpdateProductPriceHistory(productId string, unit string, price ProductPrice) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeUpdateProductPrice,
		Description: "แก้ไขราคาสินค้า " + price.CustomerType,
		Unit:        unit,
		CreatedBy:   price.UpdatedBy,
	}
}

func RemoveProductPriceHistory(productId string, unit string, price *entities.ProductPrice, createdBy string) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeRemoveProductPrice,
		Description: "ลบราคาสินค้า " + price.CustomerType + " " + unit,
		Unit:        unit,
		CreatedBy:   createdBy,
	}
}

func AddProductStockHistory(productId string, unit string, stock ProductStock, balance int) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeAddProductStock,
		Description: "เพิ่มสต็อกสินค้า " + stock.LotNumber + " จำนวน " + strconv.Itoa(stock.Quantity) + " " + unit,
		Unit:        unit,
		Import:      stock.Quantity,
		Quantity:    stock.Quantity,
		CostPrice:   stock.CostPrice,
		Price:       stock.Price,
		Balance:     balance,
		CreatedBy:   stock.UpdatedBy,
	}
}

func UpdateProductStockHistory(productId string, unit string, stock UpdateProductStock, balance int) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeUpdateProductStock,
		Description: "แก้ไขสต็อกสินค้า " + stock.LotNumber,
		Unit:        unit,
		CostPrice:   stock.CostPrice,
		Price:       stock.Price,
		Balance:     balance,
		CreatedBy:   stock.UpdatedBy,
	}
}

func RemoveProductStockHistory(productId string, unit string, stock *entities.ProductStock, balance int, createdBy string) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeRemoveProductStock,
		Description: "ลบสต็อกสินค้า " + stock.LotNumber + " จำนวน " + strconv.Itoa(stock.Quantity) + " " + unit,
		Unit:        unit,
		Balance:     balance,
		CreatedBy:   createdBy,
	}
}

func UpdateProductStockQuantityHistory(productId string, unit string, stock UpdateProductStockQuantity, balance int) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeUpdateProductStockQuantity,
		Description: "แก้ไขจำนวนสต็อกสินค้า" + " จำนวน " + strconv.Itoa(stock.Quantity) + " " + unit,
		Unit:        unit,
		Quantity:    stock.Quantity,
		Balance:     balance,
		CreatedBy:   stock.UpdatedBy,
	}
}

func AddOrderItemProductHistory(productId string, unit string, item OrderItem, balance int, createdBy string) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeAddOrderItemProduct,
		Description: "ขายสินค้า" + " จำนวน " + strconv.Itoa(item.Quantity) + " " + unit,
		Unit:        unit,
		Quantity:    item.Quantity,
		CostPrice:   item.CostPrice,
		Price:       item.Price,
		Balance:     balance,
		CreatedBy:   createdBy,
	}
}

func RemoveOrderItemProductHistory(productId string, unit string, item *entities.OrderItemProductDetail, balance int, createdBy string) ProductHistory {
	return ProductHistory{
		ProductId:   productId,
		Type:        constant.HistoryTypeRemoveOrderItemProduct,
		Description: "ยกเลิกขายสินค้า" + " จำนวน " + strconv.Itoa(item.Quantity) + " " + unit,
		Unit:        unit,
		Quantity:    item.Quantity,
		CostPrice:   item.CostPrice,
		Price:       item.Price,
		Balance:     balance,
		CreatedBy:   createdBy,
	}
}
