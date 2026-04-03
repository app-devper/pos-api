package repositories

import (
	"context"
	"errors"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productEntity struct {
	client             *mongo.Client
	productsRepo       *mongo.Collection
	productPricesRepo  *mongo.Collection
	productLotsRepo    *mongo.Collection
	productUnitsRepo   *mongo.Collection
	productStockRepo   *mongo.Collection
	productHistoryRepo *mongo.Collection
	receiveRepo        *mongo.Collection
	receiveItemsRepo   *mongo.Collection
}

type IProduct interface {

	// Product
	GetProductAll(param request.GetProduct) ([]entities.ProductDetail, error)
	GetProductBySerialNumber(serialNumber string) (*entities.Product, error)
	GetProductById(id string) (*entities.Product, error)
	GetProductsByIds(ids []string) ([]entities.Product, error)
	CreateProduct(param request.Product) (*entities.Product, error)
	CreateProductCatalog(param request.Product) (*entities.Product, error)
	CreateProductReceive(param request.Product) (*entities.Product, error)
	RemoveProductById(id string) (*entities.Product, error)
	UpdateProductById(id string, param request.UpdateProduct) (*entities.Product, error)
	RemoveQuantitySoldFirstById(id string, quantity int) (*entities.Product, error)
	AddQuantitySoldFirstById(id string, quantity int) (*entities.Product, error)
	ClearQuantitySoldFirstById(id string) (*entities.Product, error)

	// ProductLot
	GetProductLotsByProductId(productId string) ([]entities.ProductLot, error)
	GetProductLotsExpireNotify(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error)
	GetProductLots(param request.GetProductLotsExpireRange) ([]entities.ProductLot, error)
	GetProductLotById(id string) (*entities.ProductLot, error)
	CreateProductLot(param request.ProductLot) (*entities.ProductLot, error)
	UpdateProductLotById(id string, param request.UpdateProductLot) (*entities.ProductLot, error)
	RemoveProductLotById(id string) (*entities.ProductLot, error)

	// ProductUnit
	CreateProductUnitCascade(param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error)
	CreateProductUnit(param request.ProductUnit) (*entities.ProductUnit, error)
	GetProductUnitById(id string) (*entities.ProductUnit, error)
	GetProductUnitByDefault(productId string, unit string) (*entities.ProductUnit, error)
	GetProductUnitByUnit(productId string, unit string) (*entities.ProductUnit, error)
	UpdateProductUnitById(id string, param request.ProductUnit) (*entities.ProductUnit, error)
	RemoveProductUnitById(id string) (*entities.ProductUnit, error)
	RemoveProductUnitCascade(id string, branchId string, userId string) (*entities.ProductUnit, error)
	GetProductUnitsByProductId(productId string) ([]entities.ProductUnit, error)

	// ProductPrice
	GetProductPricesByProductId(productId string) ([]entities.ProductPrice, error)
	CreateProductPriceCascade(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error)
	CreateProductPrice(param request.ProductPrice) (*entities.ProductPrice, error)
	RemoveProductPriceCascade(id string, branchId string, userId string) (*entities.ProductPrice, error)
	RemoveProductPriceById(id string) (*entities.ProductPrice, error)
	RemoveProductPricesByUnitId(unitId string) error
	UpdateProductPriceById(id string, param request.ProductPrice) (*entities.ProductPrice, error)

	// ProductStock
	CreateProductStock(param request.ProductStock) (*entities.ProductStock, error)
	GetProductStockById(id string) (*entities.ProductStock, error)
	UpdateProductStockById(id string, param request.UpdateProductStock) (*entities.ProductStock, error)
	UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error)
	UpdateProductStockSequence(param request.UpdateProductStockSequence) ([]entities.ProductStock, error)
	RemoveProductStockById(id string) (*entities.ProductStock, error)
	GetProductStocksByProductId(productId string, branchId string) ([]entities.ProductStock, error)
	GetProductStockMaxSequence(productId string, unitId string) int
	GetProductStockBalance(productId string, unitId string, branchId string) int
	RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)
	AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)

	// ProductHistory
	CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error)
	RemoveProductHistoryById(id string) (*entities.ProductHistory, error)
	GetProductHistoryByProductId(productId string, branchId string) ([]entities.ProductHistory, error)
	GetProductHistoryByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.ProductHistory, error)

	// Reports
	GetLowStockProducts(threshold int, branchId string) ([]entities.LowStockProduct, error)
	GetStockReport(branchId string) ([]entities.StockReport, error)
	GetDeadStockProducts(days int, branchId string) ([]entities.DeadStockProduct, error)
}

func NewProductEntity(resource *db.Resource) IProduct {
	productsRepo := resource.PosDb.Collection("products")
	productPricesRepo := resource.PosDb.Collection("product_prices")
	productUnitsRepo := resource.PosDb.Collection("product_units")
	productLotsRepo := resource.PosDb.Collection("product_lots")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	productHistoryRepo := resource.PosDb.Collection("product_histories")
	receiveRepo := resource.PosDb.Collection("receives")
	receiveItemsRepo := resource.PosDb.Collection("receive_items")
	entity := &productEntity{
		client:             resource.Client,
		productsRepo:       productsRepo,
		productPricesRepo:  productPricesRepo,
		productLotsRepo:    productLotsRepo,
		productUnitsRepo:   productUnitsRepo,
		productStockRepo:   productStockRepo,
		productHistoryRepo: productHistoryRepo,
		receiveRepo:        receiveRepo,
		receiveItemsRepo:   receiveItemsRepo,
	}
	ensureProductIndexes(productStockRepo, productHistoryRepo)
	ensureProductCollectionIndexes(productsRepo, productUnitsRepo, productPricesRepo, productLotsRepo)
	return entity
}

func ensureProductIndexes(productStockRepo *mongo.Collection, productHistoryRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := productStockRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_stocks branchId+productId index: ", err)
	}

	_, err = productHistoryRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	if err != nil {
		logrus.Error("failed to create product_histories branchId+createdDate index: ", err)
	}

	_, err = productHistoryRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_histories productId index: ", err)
	}
}

func ensureProductCollectionIndexes(productsRepo, productUnitsRepo, productPricesRepo, productLotsRepo *mongo.Collection) {
	ctx, cancel := utils.InitContext()
	defer cancel()

	_, err := productsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "serialNumber", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logrus.Error("failed to create products serialNumber index: ", err)
	}

	_, err = productUnitsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_units productId index: ", err)
	}

	_, err = productPricesRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_prices productId index: ", err)
	}

	_, err = productLotsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_lots productId index: ", err)
	}

	_, err = productLotsRepo.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expireDate", Value: 1}},
	})
	if err != nil {
		logrus.Error("failed to create product_lots expireDate index: ", err)
	}
}

func toEntityDrugInfo(req *request.RequestDrugInfo) *entities.DrugInfo {
	if req == nil {
		return nil
	}
	return &entities.DrugInfo{
		GenericName:       req.GenericName,
		DrugType:          req.DrugType,
		DosageForm:        req.DosageForm,
		Strength:          req.Strength,
		Indication:        req.Indication,
		Dosage:            req.Dosage,
		SideEffects:       req.SideEffects,
		Contraindications: req.Contraindications,
		StorageCondition:  req.StorageCondition,
		Manufacturer:      req.Manufacturer,
		RegistrationNo:    req.RegistrationNo,
		IsControlled:      req.IsControlled,
		DrugInteractions:  req.DrugInteractions,
	}
}

func (entity *productEntity) GetProductAll(param request.GetProduct) (items []entities.ProductDetail, err error) {
	logrus.Info("GetProductAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	query := bson.M{"deletedDate": bson.M{"$exists": false}}
	if param.Category != "" {
		query["category"] = param.Category
	}

	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$lookup": bson.M{
				"from":         "product_units",
				"localField":   "_id",
				"foreignField": "productId",
				"as":           "units",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "product_prices",
				"localField":   "_id",
				"foreignField": "productId",
				"as":           "prices",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "product_stocks",
				"localField":   "_id",
				"foreignField": "productId",
				"as":           "stocks",
			},
		},
	}
	cursor, err := entity.productsRepo.Aggregate(ctx, pipeline)

	if err != nil {
		return nil, err
	}
	items = []entities.ProductDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductBySerialNumber(serialNumber string) (*entities.Product, error) {
	logrus.Info("GetProductBySerialNumber")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductBySerialNumberWithContext(ctx, serialNumber)
}

func (entity *productEntity) getProductBySerialNumberWithContext(ctx context.Context, serialNumber string) (*entities.Product, error) {
	var data entities.Product
	err := entity.productsRepo.FindOne(ctx, bson.M{"serialNumber": serialNumber, "deletedDate": bson.M{"$exists": false}}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProduct(param request.Product) (*entities.Product, error) {
	logrus.Info("CreateProduct")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductWithContext(ctx, param)
}

func (entity *productEntity) createProductWithContext(ctx context.Context, param request.Product) (*entities.Product, error) {
	serialNumber := strings.TrimSpace(param.SerialNumber)
	data := entities.Product{}
	data.Id = primitive.NewObjectID()
	data.Name = param.Name
	data.NameEn = param.NameEn
	data.Description = param.Description
	data.SerialNumber = serialNumber
	data.Unit = param.Unit
	data.Price = param.Price
	data.CostPrice = param.CostPrice
	data.Quantity = param.Quantity
	data.Category = param.Category
	data.Status = param.Status
	data.MinStock = param.MinStock
	data.DrugInfo = toEntityDrugInfo(param.DrugInfo)
	data.DrugRegistrations = param.DrugRegistrations
	data.CreatedBy = param.CreatedBy
	data.CreatedDate = time.Now()
	data.UpdatedBy = param.CreatedBy
	data.UpdatedDate = time.Now()
	_, err := entity.productsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductReceive(param request.Product) (*entities.Product, error) {
	logrus.Info("CreateProductReceive")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createProductReceiveWithContext(ctx, param)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Product
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.createProductReceiveWithContext(sessCtx, param)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) CreateProductCatalog(param request.Product) (*entities.Product, error) {
	logrus.Info("CreateProductCatalog")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createProductCatalogWithContext(ctx, param)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.Product
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.createProductCatalogWithContext(sessCtx, param)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) createProductReceiveWithContext(ctx context.Context, param request.Product) (*entities.Product, error) {
	product, err := entity.getProductBySerialNumberWithContext(ctx, strings.TrimSpace(param.SerialNumber))
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	if product != nil {
		updateProduct := request.UpdateProduct{
			Description:       param.Description,
			Category:          param.Category,
			Name:              param.Name,
			NameEn:            param.NameEn,
			Status:            param.Status,
			MinStock:          param.MinStock,
			DrugInfo:          param.DrugInfo,
			DrugRegistrations: param.DrugRegistrations,
			UpdatedBy:         param.CreatedBy,
		}
		product, err = entity.updateProductByIdWithContext(ctx, product.Id.Hex(), updateProduct)
		if err != nil {
			return nil, err
		}
	} else {
		product, err = entity.createProductWithContext(ctx, param)
		if err != nil {
			return nil, err
		}
		addHistory := request.AddProductHistory(product.Id.Hex(), param)
		addHistory.BranchId = param.BranchId
		if _, err = entity.createProductHistoryWithContext(ctx, addHistory); err != nil {
			return nil, err
		}
	}

	unit, err := entity.getProductUnitByDefaultWithContext(ctx, product.Id.Hex(), param.Unit)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if unit == nil {
		productUnit := request.ProductUnit{
			ProductId: product.Id.Hex(),
			Unit:      param.Unit,
			Size:      1,
			CostPrice: param.CostPrice,
			Barcode:   param.SerialNumber,
			UpdatedBy: param.CreatedBy,
		}
		unit, err = entity.createProductUnitWithContext(ctx, productUnit)
		if err != nil {
			return nil, err
		}

		productPrice := request.ProductPrice{
			ProductId:    product.Id.Hex(),
			UnitId:       unit.Id.Hex(),
			Price:        param.Price,
			CustomerType: constant.CustomerTypeGeneral,
			UpdatedBy:    param.CreatedBy,
		}
		if _, err = entity.createProductPriceWithContext(ctx, productPrice); err != nil {
			return nil, err
		}
	}

	if param.ReceiveId != "" {
		receive, err := entity.getReceiveByIDWithContext(ctx, param.ReceiveId)
		if err != nil {
			return nil, err
		}
		param.ReceiveCode = receive.Code
		if _, err = entity.createReceiveItemWithContext(ctx, param.ReceiveId, product.Id.Hex(), param); err != nil {
			return nil, err
		}
	}

	if param.Quantity <= 0 {
		return product, nil
	}

	unit, err = entity.getProductUnitByUnitWithContext(ctx, product.Id.Hex(), param.Unit)
	if err != nil {
		return nil, err
	}

	productStock := request.ProductStock{
		ProductId:   product.Id.Hex(),
		UnitId:      unit.Id.Hex(),
		ReceiveCode: param.ReceiveCode,
		Quantity:    param.Quantity,
		Price:       0,
		CostPrice:   0,
		ExpireDate:  param.ExpireDate,
		LotNumber:   param.LotNumber,
		ImportDate:  time.Now(),
		UpdatedBy:   param.CreatedBy,
		BranchId:    param.BranchId,
	}
	stock, err := entity.createProductStockWithContext(ctx, productStock)
	if err != nil {
		return nil, err
	}

	balance, err := entity.getProductStockBalanceWithContext(ctx, stock.ProductId.Hex(), stock.UnitId.Hex(), param.BranchId)
	if err != nil {
		return nil, err
	}
	stockHistory := request.AddProductStockHistory(stock.ProductId.Hex(), param.Unit, productStock, balance)
	stockHistory.BranchId = param.BranchId
	if _, err = entity.createProductHistoryWithContext(ctx, stockHistory); err != nil {
		return nil, err
	}

	return product, nil
}

func (entity *productEntity) createProductCatalogWithContext(ctx context.Context, param request.Product) (*entities.Product, error) {
	product, err := entity.getProductBySerialNumberWithContext(ctx, strings.TrimSpace(param.SerialNumber))
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	if product != nil {
		updateProduct := request.UpdateProduct{
			Description:       param.Description,
			Category:          param.Category,
			Name:              param.Name,
			NameEn:            param.NameEn,
			Status:            param.Status,
			MinStock:          param.MinStock,
			DrugInfo:          param.DrugInfo,
			DrugRegistrations: param.DrugRegistrations,
			UpdatedBy:         param.CreatedBy,
		}
		product, err = entity.updateProductByIdWithContext(ctx, product.Id.Hex(), updateProduct)
		if err != nil {
			return nil, err
		}
		if param.BranchId != "" {
			updHistory := request.UpdateProductHistory(product.Id.Hex(), updateProduct)
			updHistory.BranchId = param.BranchId
			if _, err = entity.createProductHistoryWithContext(ctx, updHistory); err != nil {
				return nil, err
			}
		}
	} else {
		product, err = entity.createProductWithContext(ctx, param)
		if err != nil {
			return nil, err
		}
		if param.BranchId != "" {
			addHistory := request.AddProductHistory(product.Id.Hex(), param)
			addHistory.BranchId = param.BranchId
			if _, err = entity.createProductHistoryWithContext(ctx, addHistory); err != nil {
				return nil, err
			}
		}
	}

	unit, err := entity.getProductUnitByDefaultWithContext(ctx, product.Id.Hex(), param.Unit)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if unit != nil {
		return product, nil
	}

	productUnit := request.ProductUnit{
		ProductId: product.Id.Hex(),
		Unit:      param.Unit,
		Size:      1,
		CostPrice: param.CostPrice,
		Barcode:   param.SerialNumber,
		UpdatedBy: param.CreatedBy,
	}
	unit, err = entity.createProductUnitWithContext(ctx, productUnit)
	if err != nil {
		return nil, err
	}

	productPrice := request.ProductPrice{
		ProductId:    product.Id.Hex(),
		UnitId:       unit.Id.Hex(),
		Price:        param.Price,
		CustomerType: constant.CustomerTypeGeneral,
		UpdatedBy:    param.CreatedBy,
	}
	if _, err = entity.createProductPriceWithContext(ctx, productPrice); err != nil {
		return nil, err
	}

	return product, nil
}

func (entity *productEntity) GetProductById(id string) (*entities.Product, error) {
	logrus.Info("GetProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.Product
	err = entity.productsRepo.FindOne(ctx, bson.M{"_id": objId, "deletedDate": bson.M{"$exists": false}}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductsByIds(ids []string) ([]entities.Product, error) {
	logrus.Info("GetProductsByIds")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objIds := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objId, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			objIds = append(objIds, objId)
		}
	}
	var items []entities.Product
	cursor, err := entity.productsRepo.Find(ctx, bson.M{"_id": bson.M{"$in": objIds}})
	if err != nil {
		return nil, err
	}
	items = []entities.Product{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) RemoveProductById(id string) (*entities.Product, error) {
	logrus.Info("RemoveProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx,
		bson.M{"_id": objId, "deletedDate": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"deletedDate": now}},
		opts,
	).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductById(id string, param request.UpdateProduct) (*entities.Product, error) {
	logrus.Info("UpdateProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.updateProductByIdWithContext(ctx, id, param)
}

func (entity *productEntity) updateProductByIdWithContext(ctx context.Context, id string, param request.UpdateProduct) (*entities.Product, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"name":              param.Name,
		"nameEn":            param.NameEn,
		"description":       param.Description,
		"category":          param.Category,
		"status":            param.Status,
		"minStock":          param.MinStock,
		"drugInfo":          toEntityDrugInfo(param.DrugInfo),
		"drugRegistrations": param.DrugRegistrations,
		"updatedBy":         param.UpdatedBy,
		"updatedDate":       time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveQuantityById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("RemoveQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{"quantity": -quantity},
		"$set": bson.M{"updatedDate": time.Now()},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) AddQuantityById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("AddQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{"quantity": quantity},
		"$set": bson.M{"updatedDate": time.Now()},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetTotalCostPrice(id string, quantity int) float64 {
	logrus.Info("GetTotalCostPrice")
	data, err := entity.GetProductById(id)
	if err != nil {
		return 0
	}
	return data.CostPrice * float64(quantity)
}

func (entity *productEntity) RemoveQuantitySoldFirstById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("RemoveQuantitySoldFirstById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{"soldFirst": -quantity},
		"$set": bson.M{"updatedDate": time.Now()},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) AddQuantitySoldFirstById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("AddQuantitySoldFirstById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{"soldFirst": quantity},
		"$set": bson.M{"updatedDate": time.Now()},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) ClearQuantitySoldFirstById(id string) (*entities.Product, error) {
	logrus.Info("ClearQuantitySoldFirstById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.Product
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"soldFirst":   0,
		"updatedDate": time.Now(),
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductLotByProductId(productId string, param request.Product) (*entities.ProductLot, error) {
	logrus.Info("CreateProductLotByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductLot{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(productId)
	data.LotNumber = param.LotNumber
	data.ExpireDate = param.ExpireDate
	data.Quantity = param.Quantity
	data.CostPrice = param.CostPrice
	data.CreatedBy = param.CreatedBy
	data.Notify = true
	data.UpdatedBy = param.CreatedBy
	data.CreatedDate = time.Now()
	data.UpdatedDate = time.Now()

	_, err := entity.productLotsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductLot(param request.ProductLot) (*entities.ProductLot, error) {
	logrus.Info("CreateProductLot")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductLot{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.LotNumber = param.LotNumber
	data.ExpireDate = param.ExpireDate
	data.Quantity = param.Quantity
	data.CostPrice = param.CostPrice
	data.CreatedBy = param.UpdatedBy
	data.Notify = true
	data.UpdatedBy = param.UpdatedBy
	data.CreatedDate = time.Now()
	data.UpdatedDate = time.Now()

	_, err := entity.productLotsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductLots(param request.GetProductLotsExpireRange) (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLots")
	ctx, cancel := utils.InitContext()
	defer cancel()
	filter := bson.M{}
	if !param.StartDate.IsZero() && !param.EndDate.IsZero() {
		filter["expireDate"] = bson.M{
			"$gt": param.StartDate,
			"$lt": param.EndDate,
		}
	}
	opts := options.Find().SetSort(bson.D{{Key: "expireDate", Value: -1}})
	cursor, err := entity.productLotsRepo.Find(ctx, filter, opts)

	if err != nil {
		return nil, err
	}
	items = []entities.ProductLot{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsByProductId(productId string) (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLotsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productLotsRepo.Find(ctx, bson.M{"productId": objId})
	if err != nil {
		return nil, err
	}
	items = []entities.ProductLot{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) RemoveProductLotById(id string) (*entities.ProductLot, error) {
	logrus.Info("RemoveProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductLot{}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.productLotsRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductLotsByIds(ids []string) (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLotsByIds")
	ctx, cancel := utils.InitContext()
	defer cancel()

	objIds := make([]primitive.ObjectID, 0, len(ids))
	for _, value := range ids {
		id, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return nil, err
		}
		objIds = append(objIds, id)
	}

	cursor, err := entity.productLotsRepo.Find(ctx, bson.M{"_id": bson.M{"$in": objIds}})
	if err != nil {
		return nil, err
	}
	items = []entities.ProductLot{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpired() (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLotsExpired")
	ctx, cancel := utils.InitContext()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "expireDate", Value: -1}})
	cursor, err := entity.productLotsRepo.Find(ctx,
		bson.M{"expireDate": bson.M{"$lte": time.Now()}},
		opts,
	)
	if err != nil {
		return nil, err
	}
	items = []entities.ProductLot{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpireNotify(param request.GetProductLotsExpireRange) (items []entities.ProductLotDetail, err error) {
	logrus.Info("GetProductLotsExpireNotify")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cursor, err := entity.productLotsRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"expireDate": bson.M{
					"$gte": param.StartDate,
					"$lt":  param.EndDate,
				},
				"notify": true,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "productId",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		{"$unwind": "$product"},
	})

	if err != nil {
		return nil, err
	}
	items = []entities.ProductLotDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductLotById(id string) (*entities.ProductLot, error) {
	logrus.Info("GetProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductLot{}
	err := entity.productLotsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductLotById(id string, param request.UpdateProductLot) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductLot
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"lotNumber":   param.LotNumber,
		"expireDate":  param.ExpireDate,
		"quantity":    param.Quantity,
		"costPrice":   param.CostPrice,
		"updatedDate": time.Now(),
		"updatedBy":   param.UpdatedBy,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductLotNotifyById(id string, param request.UpdateProductLotNotify) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotNotifyById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductLot
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"notify":      param.Notify,
		"updatedDate": time.Now(),
		"updatedBy":   param.UpdatedBy,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductLotQuantityById(id string, param request.UpdateProductLotQuantity) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductLot
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"quantity":    param.Quantity,
		"updatedDate": time.Now(),
		"updatedBy":   param.UpdatedBy,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductUnitByProductId(productId string, param request.Product) (*entities.ProductUnit, error) {
	logrus.Info("CreateProductUnitByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductUnit{}
	product, _ := primitive.ObjectIDFromHex(productId)
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": product, "unit": param.Unit}).Decode(&data)
	if err != nil {
		data.Id = primitive.NewObjectID()
		data.ProductId, _ = primitive.ObjectIDFromHex(productId)
		data.Unit = param.Unit
		data.Size = 1
		data.CostPrice = param.CostPrice
		data.Volume = 0
		data.VolumeUnit = ""
		data.Barcode = param.SerialNumber
		_, err = entity.productUnitsRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	} else {
		data.CostPrice = param.CostPrice
		data.Barcode = param.SerialNumber

		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err = entity.productUnitsRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}

func (entity *productEntity) CreateProductStockByProductAndUnitId(productId string, unitId string, param request.Product) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStockByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := entities.ProductStock{}
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)

	data.Id = primitive.NewObjectID()
	data.BranchId, _ = primitive.ObjectIDFromHex(param.BranchId)
	data.ProductId = product
	data.UnitId = unit
	data.Sequence = entity.GetProductStockMaxSequence(productId, unitId) + 1
	data.LotNumber = param.LotNumber
	data.CostPrice = param.CostPrice
	data.Price = param.Price
	data.Import = param.Quantity
	data.Quantity = param.Quantity
	data.ExpireDate = param.ExpireDate
	data.ImportDate = time.Now()
	data.ReceiveCode = param.ReceiveCode

	_, err := entity.productStockRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductStocksByProductAndUnitId(productId string, unitId string) (items []entities.ProductStock, err error) {
	logrus.Info("GetProductStocksByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)
	opts := options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}})
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"productId": product, "unitId": unit}, opts)
	if err != nil {
		return nil, err
	}
	items = []entities.ProductStock{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductStockMaxSequence(productId string, unitId string) int {
	logrus.Info("GetProductStockMaxSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductStockMaxSequenceWithContext(ctx, productId, unitId)
}

func (entity *productEntity) getProductStockMaxSequenceWithContext(ctx context.Context, productId string, unitId string) int {
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)
	opts := options.FindOne().SetSort(bson.D{{Key: "sequence", Value: -1}})
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"productId": product, "unitId": unit}, opts).Decode(&data)
	if err != nil {
		return 0
	}
	return data.Sequence
}

func (entity *productEntity) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	logrus.Info("AddProductStockQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(stockId)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductStock
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{"quantity": quantity},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	logrus.Info("RemoveProductStockQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(stockId)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductStock
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{
		"_id":      objId,
		"quantity": bson.M{"$gte": quantity},
	}, bson.M{
		"$inc": bson.M{"quantity": -quantity},
	}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductUnitsByProductId(productId string) (items []entities.ProductUnit, err error) {
	logrus.Info("GetProductUnitsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productUnitsRepo.Find(ctx, bson.M{"productId": product})
	if err != nil {
		return nil, err
	}
	items = []entities.ProductUnit{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) GetProductPricesByProductId(productId string) (items []entities.ProductPrice, err error) {
	logrus.Info("GetProductPricesByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productPricesRepo.Find(ctx, bson.M{"productId": product})
	if err != nil {
		return nil, err
	}
	items = []entities.ProductPrice{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) CreateProductPrice(param request.ProductPrice) (*entities.ProductPrice, error) {
	logrus.Info("CreateProductPrice")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductPriceWithContext(ctx, param)
}

func (entity *productEntity) CreateProductPriceCascade(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error) {
	logrus.Info("CreateProductPriceCascade")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createProductPriceCascadeWithContext(ctx, param, branchId, userId)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.ProductPrice
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.createProductPriceCascadeWithContext(sessCtx, param, branchId, userId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) createProductPriceWithContext(ctx context.Context, param request.ProductPrice) (*entities.ProductPrice, error) {
	data := entities.ProductPrice{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.UnitId, _ = primitive.ObjectIDFromHex(param.UnitId)
	data.CustomerType = param.CustomerType
	data.Price = param.Price
	_, err := entity.productPricesRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductPriceById(id string, param request.ProductPrice) (*entities.ProductPrice, error) {
	logrus.Info("UpdateProductPriceById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductPrice
	err = entity.productPricesRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"customerType": param.CustomerType,
		"price":        param.Price,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductPriceById(id string) (*entities.ProductPrice, error) {
	logrus.Info("RemoveProductPriceById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.removeProductPriceByIDWithContext(ctx, id)
}

func (entity *productEntity) RemoveProductPriceCascade(id string, branchId string, userId string) (*entities.ProductPrice, error) {
	logrus.Info("RemoveProductPriceCascade")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.removeProductPriceCascadeWithContext(ctx, id, branchId, userId)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.ProductPrice
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.removeProductPriceCascadeWithContext(sessCtx, id, branchId, userId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) removeProductPriceByIDWithContext(ctx context.Context, id string) (*entities.ProductPrice, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.ProductPrice
	err = entity.productPricesRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	if data.CustomerType == constant.CustomerTypeGeneral {
		return nil, errors.New("can not remove default price")
	}
	err = entity.productPricesRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductPricesByUnitId(unitId string) error {
	logrus.Info("RemoveProductPricesByUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(unitId)
	if err != nil {
		return err
	}
	_, err = entity.productPricesRepo.DeleteMany(ctx, bson.M{"unitId": objId})
	return err

}

func (entity *productEntity) CreateProductUnit(param request.ProductUnit) (*entities.ProductUnit, error) {
	logrus.Info("CreateProductUnit")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductUnitWithContext(ctx, param)
}

func (entity *productEntity) createProductUnitWithContext(ctx context.Context, param request.ProductUnit) (*entities.ProductUnit, error) {
	data := entities.ProductUnit{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.Unit = param.Unit
	data.Size = param.Size
	data.CostPrice = param.CostPrice
	data.Volume = param.Volume
	data.VolumeUnit = param.VolumeUnit
	data.Barcode = param.Barcode
	_, err := entity.productUnitsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	logrus.Info("GetProductUnitById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductUnitByIDWithContext(ctx, id)
}

func (entity *productEntity) getProductUnitByIDWithContext(ctx context.Context, id string) (*entities.ProductUnit, error) {
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductUnit{}
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductUnitByDefault(productId string, unit string) (*entities.ProductUnit, error) {
	logrus.Info("GetProductUnitByDefault")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductUnitByDefaultWithContext(ctx, productId, unit)
}

func (entity *productEntity) getProductUnitByDefaultWithContext(ctx context.Context, productId string, unit string) (*entities.ProductUnit, error) {
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	data := entities.ProductUnit{}
	err = entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": product, "unit": unit, "size": 1}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductUnitByUnit(productId string, unit string) (*entities.ProductUnit, error) {
	logrus.Info("GetProductUnitByUnit")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductUnitByUnitWithContext(ctx, productId, unit)
}

func (entity *productEntity) getProductUnitByUnitWithContext(ctx context.Context, productId string, unit string) (*entities.ProductUnit, error) {
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	data := entities.ProductUnit{}
	err = entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": product, "unit": unit}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductUnitById(id string, param request.ProductUnit) (*entities.ProductUnit, error) {
	logrus.Info("UpdateProductUnitById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductUnit
	err = entity.productUnitsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"unit":       param.Unit,
		"size":       param.Size,
		"costPrice":  param.CostPrice,
		"volume":     param.Volume,
		"volumeUnit": param.VolumeUnit,
		"barcode":    param.Barcode,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductUnitById(id string) (*entities.ProductUnit, error) {
	logrus.Info("RemoveProductUnitById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.removeProductUnitByIDWithContext(ctx, id)
}

func (entity *productEntity) RemoveProductUnitCascade(id string, branchId string, userId string) (*entities.ProductUnit, error) {
	logrus.Info("RemoveProductUnitCascade")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.removeProductUnitCascadeWithContext(ctx, id, branchId, userId)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.ProductUnit
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.removeProductUnitCascadeWithContext(sessCtx, id, branchId, userId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) CreateProductUnitCascade(param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error) {
	logrus.Info("CreateProductUnitCascade")
	ctx, cancel := utils.InitContext()
	defer cancel()

	if entity.client == nil {
		return entity.createProductUnitCascadeWithContext(ctx, param, branchId, userId)
	}

	session, err := entity.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var result *entities.ProductUnit
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		data, txErr := entity.createProductUnitCascadeWithContext(sessCtx, param, branchId, userId)
		if txErr != nil {
			return nil, txErr
		}
		result = data
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (entity *productEntity) removeProductUnitByIDWithContext(ctx context.Context, id string) (*entities.ProductUnit, error) {
	data := entities.ProductUnit{}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.productUnitsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	if data.Size == 1 {
		return nil, errors.New("can not remove default unit")
	}
	err = entity.productUnitsRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) removeProductUnitCascadeWithContext(ctx context.Context, id string, branchId string, userId string) (*entities.ProductUnit, error) {
	unit, err := entity.removeProductUnitByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}

	unitObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	if _, err = entity.productPricesRepo.DeleteMany(ctx, bson.M{"unitId": unitObjID}); err != nil {
		return nil, err
	}

	if branchId != "" {
		history := request.RemoveProductUnitHistory(unit.ProductId.Hex(), unit, userId)
		history.BranchId = branchId
		if _, err = entity.createProductHistoryWithContext(ctx, history); err != nil {
			return nil, err
		}
	}

	return unit, nil
}

func (entity *productEntity) createProductUnitCascadeWithContext(ctx context.Context, param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error) {
	productUnit := request.ProductUnit{
		ProductId:  param.ProductId,
		Unit:       param.Unit,
		Size:       param.Size,
		CostPrice:  param.CostPrice,
		Barcode:    param.Barcode,
		Volume:     param.Volume,
		VolumeUnit: param.VolumeUnit,
		UpdatedBy:  userId,
	}
	unit, err := entity.createProductUnitWithContext(ctx, productUnit)
	if err != nil {
		return nil, err
	}

	productPrice := request.ProductPrice{
		ProductId:    param.ProductId,
		UnitId:       unit.Id.Hex(),
		Price:        param.Price,
		CustomerType: constant.CustomerTypeGeneral,
		UpdatedBy:    userId,
	}
	if _, err = entity.createProductPriceWithContext(ctx, productPrice); err != nil {
		return nil, err
	}

	if branchId != "" {
		history := request.AddProductUnitHistory(param.ProductId, productUnit)
		history.BranchId = branchId
		if _, err = entity.createProductHistoryWithContext(ctx, history); err != nil {
			return nil, err
		}
	}

	return unit, nil
}

func (entity *productEntity) createProductPriceCascadeWithContext(ctx context.Context, param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error) {
	param.UpdatedBy = userId
	price, err := entity.createProductPriceWithContext(ctx, param)
	if err != nil {
		return nil, err
	}

	if branchId != "" {
		unit, err := entity.getProductUnitByIDWithContext(ctx, price.UnitId.Hex())
		if err != nil {
			return nil, err
		}
		history := request.AddProductPriceHistory(param.ProductId, unit.Unit, param)
		history.BranchId = branchId
		if _, err = entity.createProductHistoryWithContext(ctx, history); err != nil {
			return nil, err
		}
	}

	return price, nil
}

func (entity *productEntity) removeProductPriceCascadeWithContext(ctx context.Context, id string, branchId string, userId string) (*entities.ProductPrice, error) {
	price, err := entity.removeProductPriceByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}

	if branchId != "" {
		unit, err := entity.getProductUnitByIDWithContext(ctx, price.UnitId.Hex())
		if err != nil {
			return nil, err
		}
		history := request.RemoveProductPriceHistory(price.ProductId.Hex(), unit.Unit, price, userId)
		history.BranchId = branchId
		if _, err = entity.createProductHistoryWithContext(ctx, history); err != nil {
			return nil, err
		}
	}

	return price, nil
}

func (entity *productEntity) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStock")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductStockWithContext(ctx, param)
}

func (entity *productEntity) createProductStockWithContext(ctx context.Context, param request.ProductStock) (*entities.ProductStock, error) {
	data := entities.ProductStock{}
	data.Id = primitive.NewObjectID()
	data.BranchId, _ = primitive.ObjectIDFromHex(param.BranchId)
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.UnitId, _ = primitive.ObjectIDFromHex(param.UnitId)
	data.Sequence = entity.getProductStockMaxSequenceWithContext(ctx, param.ProductId, param.UnitId) + 1
	data.LotNumber = param.LotNumber
	data.CostPrice = param.CostPrice
	data.Price = param.Price
	data.Import = param.Quantity
	data.Quantity = param.Quantity
	data.ExpireDate = param.ExpireDate
	data.ImportDate = param.ImportDate
	data.ReceiveCode = param.ReceiveCode
	_, err := entity.productStockRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductStockById(id string) (*entities.ProductStock, error) {
	logrus.Info("GetProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductStocksByProductId(productId string, branchId string) (items []entities.ProductStock, err error) {
	logrus.Info("GetProductStocksByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	filter := bson.M{"productId": product}
	if branchId != "" {
		if branch, branchErr := primitive.ObjectIDFromHex(branchId); branchErr == nil {
			filter["branchId"] = branch
		}
	}
	opts := options.Find().SetSort(bson.D{{Key: "expireDate", Value: 1}, {Key: "sequence", Value: 1}})
	cursor, err := entity.productStockRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items = []entities.ProductStock{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (entity *productEntity) UpdateProductStockById(id string, param request.UpdateProductStock) (*entities.ProductStock, error) {
	logrus.Info("UpdateProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductStock
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"lotNumber":  param.LotNumber,
		"costPrice":  param.CostPrice,
		"price":      param.Price,
		"expireDate": param.ExpireDate,
		"importDate": param.ImportDate,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error) {
	logrus.Info("UpdateProductStockQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	var data entities.ProductStock
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{
		"quantity": quantity,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductStockById(id string) (*entities.ProductStock, error) {
	logrus.Info("RemoveProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.ProductStock
	err = entity.productStockRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductStockSequence(param request.UpdateProductStockSequence) ([]entities.ProductStock, error) {
	logrus.Info("UpdateProductStockSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()

	sequenceMap := make(map[string]int, len(param.Stocks))
	objIds := make([]primitive.ObjectID, 0, len(param.Stocks))
	for _, s := range param.Stocks {
		id, err := primitive.ObjectIDFromHex(s.StockId)
		if err != nil {
			return nil, err
		}
		objIds = append(objIds, id)
		sequenceMap[s.StockId] = s.Sequence
	}

	writes := make([]mongo.WriteModel, 0, len(objIds))
	for _, id := range objIds {
		writes = append(writes, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": bson.M{"sequence": sequenceMap[id.Hex()]}}))
	}
	if _, err := entity.productStockRepo.BulkWrite(ctx, writes); err != nil {
		return nil, err
	}

	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"_id": bson.M{"$in": objIds}},
		options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return nil, err
	}
	stocks := []entities.ProductStock{}
	if err = cursor.All(ctx, &stocks); err != nil {
		return nil, err
	}
	return stocks, nil
}

func (entity *productEntity) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	logrus.Info("CreateProductHistory")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductHistoryWithContext(ctx, param)
}

func (entity *productEntity) createProductHistoryWithContext(ctx context.Context, param request.ProductHistory) (*entities.ProductHistory, error) {
	data := entities.ProductHistory{}
	data.Id = primitive.NewObjectID()
	data.BranchId, _ = primitive.ObjectIDFromHex(param.BranchId)
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.Description = param.Description
	data.Type = param.Type
	data.Unit = param.Unit
	data.CostPrice = param.CostPrice
	data.Price = param.Price
	data.Quantity = param.Quantity
	data.Import = param.Import
	data.CreatedBy = param.CreatedBy
	data.CreatedDate = time.Now()
	data.Balance = param.Balance

	_, err := entity.productHistoryRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductHistoryById(id string) (*entities.ProductHistory, error) {
	logrus.Info("RemoveProductHistoryById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.ProductHistory{}
	err = entity.productHistoryRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductStockBalance(productId string, unitId string, branchId string) int {
	logrus.Info("GetProductStockBalance")
	ctx, cancel := utils.InitContext()
	defer cancel()
	balance, err := entity.getProductStockBalanceWithContext(ctx, productId, unitId, branchId)
	if err != nil {
		return 0
	}
	return balance
}

func (entity *productEntity) getProductStockBalanceWithContext(ctx context.Context, productId string, unitId string, branchId string) (int, error) {
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)
	match := bson.M{"productId": product, "unitId": unit}
	if branchId != "" {
		branch, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return 0, err
		}
		match["branchId"] = branch
	}
	pipeline := []bson.M{
		{"$match": match},
		{"$group": bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}},
	}
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return 0, err
	}
	if v, ok := results[0]["balance"].(int32); ok {
		return int(v), nil
	}
	if v, ok := results[0]["balance"].(int64); ok {
		return int(v), nil
	}
	return 0, nil
}

func (entity *productEntity) getReceiveByIDWithContext(ctx context.Context, id string) (*entities.Receive, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.Receive{}
	err = entity.receiveRepo.FindOne(ctx, bson.M{"_id": objID}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) createReceiveItemWithContext(ctx context.Context, receiveId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
	recvID, err := primitive.ObjectIDFromHex(receiveId)
	if err != nil {
		return nil, err
	}
	productObjID, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	data := entities.ReceiveItem{
		ReceiveId:  recvID,
		ProductId:  productObjID,
		Quantity:   form.Quantity,
		CostPrice:  form.CostPrice,
		LotNumber:  form.LotNumber,
		ExpireDate: form.ExpireDate,
	}
	_, err = entity.receiveItemsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductHistoryByProductId(productId string, branchId string) ([]entities.ProductHistory, error) {
	logrus.Info("GetProductHistoryByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	prodObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"productId": prodObjId}
	if branchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = branchObjId
	}
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.productHistoryRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.ProductHistory
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.ProductHistory{}
	}
	return results, nil
}

func (entity *productEntity) GetProductHistoryByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.ProductHistory, error) {
	logrus.Info("GetProductHistoryByDateRange")
	ctx, cancel := utils.InitContext()
	defer cancel()
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	filter := bson.M{
		"createdDate": bson.M{"$gte": startDate, "$lte": endOfDay},
	}
	if branchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(branchId)
		filter["branchId"] = branchObjId
	}
	opts := options.Find().SetSort(bson.M{"createdDate": -1})
	cursor, err := entity.productHistoryRepo.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var results []entities.ProductHistory
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.ProductHistory{}
	}
	return results, nil
}

func (entity *productEntity) GetLowStockProducts(threshold int, branchId string) ([]entities.LowStockProduct, error) {
	logrus.Info("GetLowStockProducts")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchStage := bson.M{}
	if branchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(branchId)
		matchStage["branchId"] = branchObjId
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id":        "$productId",
			"totalStock": bson.M{"$sum": "$quantity"},
		}},
		{"$match": bson.M{"totalStock": bson.M{"$lte": threshold}}},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": "$product"},
		{"$project": bson.M{
			"_id":          1,
			"totalStock":   1,
			"name":         "$product.name",
			"serialNumber": "$product.serialNumber",
			"unit":         "$product.unit",
		}},
		{"$sort": bson.M{"totalStock": 1}},
	}

	var results []entities.LowStockProduct
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.LowStockProduct{}
	}
	return results, nil
}

func (entity *productEntity) GetStockReport(branchId string) ([]entities.StockReport, error) {
	logrus.Info("GetStockReport")
	ctx, cancel := utils.InitContext()
	defer cancel()

	matchStage := bson.M{}
	if branchId != "" {
		branchObjId, _ := primitive.ObjectIDFromHex(branchId)
		matchStage["branchId"] = branchObjId
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id":        "$productId",
			"totalStock": bson.M{"$sum": "$quantity"},
			"totalCost":  bson.M{"$sum": bson.M{"$multiply": bson.A{"$quantity", "$costPrice"}}},
		}},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": "$product"},
		{"$project": bson.M{
			"_id":          1,
			"totalStock":   1,
			"totalCost":    1,
			"name":         "$product.name",
			"serialNumber": "$product.serialNumber",
			"unit":         "$product.unit",
		}},
		{"$sort": bson.M{"name": 1}},
	}

	var results []entities.StockReport
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.StockReport{}
	}
	return results, nil
}

func (entity *productEntity) GetDeadStockProducts(days int, branchId string) ([]entities.DeadStockProduct, error) {
	logrus.Info("GetDeadStockProducts")
	ctx, cancel := utils.InitContext()
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -days)

	matchFilter := bson.M{
		"deletedDate": bson.M{"$exists": false},
		"quantity":    bson.M{"$gt": 0},
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{"$lookup": bson.M{
			"from": "order_items",
			"let":  bson.M{"pid": "$_id"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$productId", "$$pid"}}}},
				bson.M{"$sort": bson.M{"createdDate": -1}},
				bson.M{"$limit": 1},
				bson.M{"$project": bson.M{"createdDate": 1}},
			},
			"as": "lastOrder",
		}},
		{"$addFields": bson.M{
			"lastSold": bson.M{"$ifNull": bson.A{
				bson.M{"$arrayElemAt": bson.A{"$lastOrder.createdDate", 0}},
				nil,
			}},
		}},
		{"$match": bson.M{
			"$or": bson.A{
				bson.M{"lastSold": nil},
				bson.M{"lastSold": bson.M{"$lt": cutoff}},
			},
		}},
		{"$project": bson.M{
			"name":      1,
			"quantity":  1,
			"costPrice": 1,
			"lastSold": bson.M{"$cond": bson.M{
				"if":   bson.M{"$eq": bson.A{"$lastSold", nil}},
				"then": "",
				"else": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$lastSold"}},
			}},
		}},
		{"$sort": bson.M{"lastSold": 1}},
	}

	var results []entities.DeadStockProduct
	cursor, err := entity.productsRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []entities.DeadStockProduct{}
	}
	return results, nil
}
