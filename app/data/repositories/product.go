package repositories

import (
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
	productsRepo       *mongo.Collection
	productPricesRepo  *mongo.Collection
	productLotsRepo    *mongo.Collection
	productUnitsRepo   *mongo.Collection
	productStockRepo   *mongo.Collection
	productHistoryRepo *mongo.Collection
}

type IProduct interface {

	// Product
	GetProductAll(param request.GetProduct) ([]entities.ProductDetail, error)
	GetProductBySerialNumber(serialNumber string) (*entities.Product, error)
	GetProductById(id string) (*entities.Product, error)
	GetProductsByIds(ids []string) ([]entities.Product, error)
	CreateProduct(param request.Product) (*entities.Product, error)
	RemoveProductById(id string) (*entities.Product, error)
	UpdateProductById(id string, param request.UpdateProduct) (*entities.Product, error)
	RemoveQuantitySoldFirstById(id string, quantity int) (*entities.Product, error)
	AddQuantitySoldFirstById(id string, quantity int) (*entities.Product, error)
	ClearQuantitySoldFirstById(id string) (*entities.Product, error)

	// ProductLot
	GetProductLotsByProductId(productId string) ([]entities.ProductLot, error)
	GetProductLotsExpireNotify(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error)

	// ProductUnit
	CreateProductUnit(param request.ProductUnit) (*entities.ProductUnit, error)
	GetProductUnitById(id string) (*entities.ProductUnit, error)
	GetProductUnitByDefault(productId string, unit string) (*entities.ProductUnit, error)
	GetProductUnitByUnit(productId string, unit string) (*entities.ProductUnit, error)
	UpdateProductUnitById(id string, param request.ProductUnit) (*entities.ProductUnit, error)
	RemoveProductUnitById(id string) (*entities.ProductUnit, error)
	GetProductUnitsByProductId(productId string) ([]entities.ProductUnit, error)

	// ProductPrice
	GetProductPricesByProductId(productId string) ([]entities.ProductPrice, error)
	CreateProductPrice(param request.ProductPrice) (*entities.ProductPrice, error)
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
	GetProductStocksByProductId(productId string) ([]entities.ProductStock, error)
	GetProductStockMaxSequence(productId string, unitId string) int
	GetProductStockBalance(productId string, unitId string) int
	RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)
	AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)

	// ProductHistory
	CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error)
	GetProductHistoryByProductId(productId string, branchId string) ([]entities.ProductHistory, error)
	GetProductHistoryByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.ProductHistory, error)

	// Reports
	GetLowStockProducts(threshold int, branchId string) ([]entities.LowStockProduct, error)
	GetStockReport(branchId string) ([]entities.StockReport, error)
}

func NewProductEntity(resource *db.Resource) IProduct {
	productsRepo := resource.PosDb.Collection("products")
	productPricesRepo := resource.PosDb.Collection("product_prices")
	productUnitsRepo := resource.PosDb.Collection("product_units")
	productLotsRepo := resource.PosDb.Collection("product_lots")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	productHistoryRepo := resource.PosDb.Collection("product_histories")
	entity := &productEntity{
		productsRepo:       productsRepo,
		productPricesRepo:  productPricesRepo,
		productLotsRepo:    productLotsRepo,
		productUnitsRepo:   productUnitsRepo,
		productStockRepo:   productStockRepo,
		productHistoryRepo: productHistoryRepo,
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
		Keys: bson.D{{Key: "serialNumber", Value: 1}},
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
	}
}

func (entity *productEntity) GetProductAll(param request.GetProduct) (items []entities.ProductDetail, err error) {
	logrus.Info("GetProductAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	query := bson.M{}
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
	var data entities.Product
	err := entity.productsRepo.FindOne(ctx, bson.M{"serialNumber": serialNumber}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProduct(param request.Product) (*entities.Product, error) {
	logrus.Info("CreateProduct")
	ctx, cancel := utils.InitContext()
	defer cancel()
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
	data.DrugInfo = toEntityDrugInfo(param.DrugInfo)
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

func (entity *productEntity) GetProductById(id string) (*entities.Product, error) {
	logrus.Info("GetProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data entities.Product
	err = entity.productsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
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
	var data entities.Product
	err = entity.productsRepo.FindOneAndDelete(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductById(id string, param request.UpdateProduct) (*entities.Product, error) {
	logrus.Info("UpdateProductById")
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
		"name":        param.Name,
		"nameEn":      param.NameEn,
		"description": param.Description,
		"category":    param.Category,
		"status":      param.Status,
		"drugInfo":    toEntityDrugInfo(param.DrugInfo),
		"updatedBy":   param.UpdatedBy,
		"updatedDate": time.Now(),
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
	opts := options.Find().SetSort(bson.D{{Key: "expireDate", Value: -1}})
	cursor, err := entity.productLotsRepo.Find(ctx,
		bson.M{"expireDate": bson.M{
			"$gt": param.StartDate,
			"$lt": param.EndDate,
		}},
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
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{
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
	product, err := primitive.ObjectIDFromHex(productId)
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
	product, err := primitive.ObjectIDFromHex(productId)
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

func (entity *productEntity) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStock")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductStock{}
	data.Id = primitive.NewObjectID()
	data.BranchId, _ = primitive.ObjectIDFromHex(param.BranchId)
	data.ProductId, _ = primitive.ObjectIDFromHex(param.ProductId)
	data.UnitId, _ = primitive.ObjectIDFromHex(param.UnitId)
	data.Sequence = entity.GetProductStockMaxSequence(param.ProductId, param.UnitId) + 1
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

func (entity *productEntity) GetProductStocksByProductId(productId string) (items []entities.ProductStock, err error) {
	logrus.Info("GetProductStocksByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	opts := options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}})
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"productId": product}, opts)
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

func (entity *productEntity) GetProductStockBalance(productId string, unitId string) int {
	logrus.Info("GetProductStockBalance")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)
	pipeline := []bson.M{
		{"$match": bson.M{"productId": product, "unitId": unit}},
		{"$group": bson.M{"_id": nil, "balance": bson.M{"$sum": "$quantity"}}},
	}
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return 0
	}
	if v, ok := results[0]["balance"].(int32); ok {
		return int(v)
	}
	if v, ok := results[0]["balance"].(int64); ok {
		return int(v)
	}
	return 0
}

func (entity *productEntity) GetProductHistoryByProductId(productId string, branchId string) ([]entities.ProductHistory, error) {
	logrus.Info("GetProductHistoryByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	prodObjId, _ := primitive.ObjectIDFromHex(productId)
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
	filter := bson.M{
		"createdDate": bson.M{"$gte": startDate, "$lte": endDate},
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
			"totalCost":  bson.M{"$sum": bson.M{"$multiply": []string{"$quantity", "$costPrice"}}},
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
