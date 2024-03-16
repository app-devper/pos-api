package repository

import (
	"errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
	"pos/app/domain/constant"
	"pos/app/domain/entities"
	"pos/app/domain/request"
	"pos/db"
	"strings"
	"time"
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
	GetProductAll(product request.GetProduct) ([]entities.ProductDetail, error)
	GetProductBySerialNumber(serialNumber string) (*entities.Product, error)
	GetProductById(id string) (*entities.Product, error)
	CreateProduct(form request.Product) (*entities.Product, error)
	RemoveProductById(id string) (*entities.Product, error)
	UpdateProductById(id string, form request.UpdateProduct) (*entities.Product, error)
	RemoveQuantityById(id string, quantity int) (*entities.Product, error)
	AddQuantityById(id string, quantity int) (*entities.Product, error)
	GetTotalCostPrice(id string, quantity int) float64

	// ProductLot
	CreateProductLotByProductId(productId string, form request.Product) (*entities.ProductLot, error)
	CreateProductLot(form request.ProductLot) (*entities.ProductLot, error)
	GetProductLots(form request.GetExpireRange) ([]entities.ProductLot, error)
	GetProductLotsByProductId(productId string) ([]entities.ProductLot, error)
	GetProductLotsByIds(ids []string) ([]entities.ProductLot, error)
	GetProductLotsExpired() ([]entities.ProductLot, error)
	GetProductLotsExpireNotify(form request.GetExpireRange) ([]entities.ProductLotDetail, error)
	GetProductLotById(id string) (*entities.ProductLot, error)
	RemoveProductLotById(id string) (*entities.ProductLot, error)
	UpdateProductLotById(id string, form request.UpdateProductLot) (*entities.ProductLot, error)
	UpdateProductLotNotifyById(id string, form request.UpdateProductLotNotify) (*entities.ProductLot, error)
	UpdateProductLotQuantityById(id string, form request.UpdateProductLotQuantity) (*entities.ProductLot, error)

	// ProductUnit
	CreateProductUnit(form request.ProductUnit) (*entities.ProductUnit, error)
	GetProductUnitById(id string) (*entities.ProductUnit, error)
	GetProductUnitByDefault(productId string, unit string) (*entities.ProductUnit, error)
	GetProductUnitByUnit(productId string, unit string) (*entities.ProductUnit, error)
	UpdateProductUnitById(id string, form request.ProductUnit) (*entities.ProductUnit, error)
	RemoveProductUnitById(id string) (*entities.ProductUnit, error)
	GetProductUnitsByProductId(productId string) ([]entities.ProductUnit, error)

	// ProductPrice
	GetProductPricesByProductId(productId string) ([]entities.ProductPrice, error)
	CreateProductPrice(form request.ProductPrice) (*entities.ProductPrice, error)
	RemoveProductPriceById(id string) (*entities.ProductPrice, error)
	RemoveProductPricesByUnitId(unitId string) error
	UpdateProductPriceById(id string, form request.ProductPrice) (*entities.ProductPrice, error)

	// ProductStock
	CreateProductStock(form request.ProductStock) (*entities.ProductStock, error)
	GetProductStockById(id string) (*entities.ProductStock, error)
	UpdateProductStockById(id string, form request.UpdateProductStock) (*entities.ProductStock, error)
	UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error)
	UpdateProductStockSequence(form request.UpdateProductStockSequence) ([]entities.ProductStock, error)
	RemoveProductStockById(id string) (*entities.ProductStock, error)
	GetProductStocksByProductId(productId string) ([]entities.ProductStock, error)
	GetProductStockMaxSequence(productId string, unitId string) int
	GetProductStockBalance(productId string, unitId string) int
	RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)
	AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error)

	// ProductHistory
	CreateProductHistory(form request.ProductHistory) (*entities.ProductHistory, error)
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
	return entity
}

func (entity *productEntity) GetProductAll(product request.GetProduct) (items []entities.ProductDetail, err error) {
	logrus.Info("GetProductAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	query := bson.M{}
	if product.Category != "" {
		query["category"] = product.Category
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
	for cursor.Next(ctx) {
		item := entities.ProductDetail{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductDetail{}
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

func (entity *productEntity) CreateProduct(form request.Product) (*entities.Product, error) {
	logrus.Info("CreateProduct")
	ctx, cancel := utils.InitContext()
	defer cancel()
	serialNumber := strings.TrimSpace(form.SerialNumber)
	data := entities.Product{}
	data.Id = primitive.NewObjectID()
	data.Name = form.Name
	data.NameEn = form.NameEn
	data.Description = form.Description
	data.SerialNumber = serialNumber
	data.Unit = form.Unit
	data.Price = form.Price
	data.CostPrice = form.CostPrice
	data.Quantity = form.Quantity
	data.CreatedBy = form.CreatedBy
	data.Category = form.Category
	data.CreatedDate = time.Now()
	data.UpdatedBy = form.CreatedBy
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
	var data entities.Product
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.productsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductById(id string) (*entities.Product, error) {
	logrus.Info("RemoveProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data entities.Product
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.productsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.productsRepo.DeleteOne(ctx, bson.M{"_id": objId})
	return &data, nil
}

func (entity *productEntity) UpdateProductById(id string, form request.UpdateProduct) (*entities.Product, error) {
	logrus.Info("UpdateProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data, err := entity.GetProductById(id)
	if err != nil {
		return nil, err
	}
	data.SerialNumber = form.SerialNumber
	data.Name = form.Name
	data.NameEn = form.NameEn
	data.Description = form.Description
	data.Price = form.Price
	data.CostPrice = form.CostPrice
	data.Unit = form.Unit
	data.Quantity = form.Quantity
	data.Category = form.Category
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) RemoveQuantityById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("RemoveQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)

	data, err := entity.GetProductById(id)
	if err != nil {
		return nil, err
	}
	data.Quantity = data.Quantity - quantity
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) AddQuantityById(id string, quantity int) (*entities.Product, error) {
	logrus.Info("AddQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data, err := entity.GetProductById(id)
	if err != nil {
		return nil, err
	}
	data.Quantity = data.Quantity + quantity
	data.UpdatedDate = time.Now()
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) GetTotalCostPrice(id string, quantity int) float64 {
	logrus.Info("GetTotalCostPrice")
	data, err := entity.GetProductById(id)
	if err != nil {
		return 0
	}
	return data.CostPrice * float64(quantity)
}

func (entity *productEntity) CreateProductLotByProductId(productId string, form request.Product) (*entities.ProductLot, error) {
	logrus.Info("CreateProductLotByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductLot{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(productId)
	data.LotNumber = form.LotNumber
	data.ExpireDate = form.ExpireDate
	data.Quantity = form.Quantity
	data.CostPrice = form.CostPrice
	data.CreatedBy = form.CreatedBy
	data.Notify = true
	data.UpdatedBy = form.CreatedBy
	data.CreatedDate = time.Now()
	data.UpdatedDate = time.Now()

	_, err := entity.productLotsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductLot(form request.ProductLot) (*entities.ProductLot, error) {
	logrus.Info("CreateProductLot")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductLot{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(form.ProductId)
	data.LotNumber = form.LotNumber
	data.ExpireDate = form.ExpireDate
	data.Quantity = form.Quantity
	data.CostPrice = form.CostPrice
	data.CreatedBy = form.UpdatedBy
	data.Notify = true
	data.UpdatedBy = form.UpdatedBy
	data.CreatedDate = time.Now()
	data.UpdatedDate = time.Now()

	_, err := entity.productLotsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductLots(form request.GetExpireRange) (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLots")
	ctx, cancel := utils.InitContext()
	defer cancel()
	opts := options.Find().SetSort(bson.D{{"expireDate", -1}})
	cursor, err := entity.productLotsRepo.Find(ctx,
		bson.M{"expireDate": bson.M{
			"$gt": form.StartDate,
			"$lt": form.EndDate,
		}},
		opts,
	)

	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		data := entities.ProductLot{}
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.ProductLot{}
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
	for cursor.Next(ctx) {
		var productLot entities.ProductLot
		err = cursor.Decode(&productLot)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, productLot)
		}
	}
	if items == nil {
		items = []entities.ProductLot{}
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
	err = entity.productLotsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.productLotsRepo.DeleteOne(ctx, bson.M{"_id": objId})
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
	for cursor.Next(ctx) {
		item := entities.ProductLot{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpired() (items []entities.ProductLot, err error) {
	logrus.Info("GetProductLotsExpired")
	ctx, cancel := utils.InitContext()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{"expireDate", -1}})
	cursor, err := entity.productLotsRepo.Find(ctx,
		bson.M{"expireDate": bson.M{"$lte": time.Now()}},
		opts,
	)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := entities.ProductLot{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpireNotify(form request.GetExpireRange) (items []entities.ProductLotDetail, err error) {
	logrus.Info("GetProductLotsExpireNotify")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cursor, err := entity.productLotsRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"expireDate": bson.M{
					"$gte": form.StartDate,
					"$lt":  form.EndDate,
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
	for cursor.Next(ctx) {
		data := entities.ProductLotDetail{}
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []entities.ProductLotDetail{}
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

func (entity *productEntity) UpdateProductLotById(id string, form request.UpdateProductLot) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data, err := entity.GetProductLotById(id)
	if err != nil {
		return nil, err
	}

	data.LotNumber = form.LotNumber
	data.ExpireDate = form.ExpireDate
	data.Quantity = form.Quantity
	data.CostPrice = form.CostPrice
	data.UpdatedDate = time.Now()
	data.UpdatedBy = form.UpdatedBy

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) UpdateProductLotNotifyById(id string, form request.UpdateProductLotNotify) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotNotifyById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data, err := entity.GetProductLotById(id)
	if err != nil {
		return nil, err
	}

	data.Notify = form.Notify
	data.UpdatedDate = time.Now()
	data.UpdatedBy = form.UpdatedBy

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) UpdateProductLotQuantityById(id string, form request.UpdateProductLotQuantity) (*entities.ProductLot, error) {
	logrus.Info("UpdateProductLotQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data, err := entity.GetProductLotById(id)
	if err != nil {
		return nil, err
	}

	data.Quantity = form.Quantity
	data.UpdatedDate = time.Now()
	data.UpdatedBy = form.UpdatedBy

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productLotsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (entity *productEntity) CreateProductUnitByProductId(productId string, form request.Product) (*entities.ProductUnit, error) {
	logrus.Info("CreateProductUnitByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductUnit{}
	product, _ := primitive.ObjectIDFromHex(productId)
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": product, "unit": form.Unit}).Decode(&data)
	if err != nil {
		data.Id = primitive.NewObjectID()
		data.ProductId, _ = primitive.ObjectIDFromHex(productId)
		data.Unit = form.Unit
		data.Size = 1
		data.CostPrice = form.CostPrice
		data.Volume = 0
		data.VolumeUnit = ""
		data.Barcode = form.SerialNumber
		_, err = entity.productUnitsRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	} else {
		data.CostPrice = form.CostPrice
		data.Barcode = form.SerialNumber

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

func (entity *productEntity) CreateProductStockByProductAndUnitId(productId string, unitId string, form request.Product) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStockByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := entities.ProductStock{}
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)

	data.Id = primitive.NewObjectID()
	data.ProductId = product
	data.UnitId = unit
	data.Sequence = entity.GetProductStockMaxSequence(productId, unitId) + 1
	data.LotNumber = form.LotNumber
	data.CostPrice = form.CostPrice
	data.Price = form.Price
	data.Import = form.Quantity
	data.Quantity = form.Quantity
	data.ExpireDate = form.ExpireDate
	data.ImportDate = time.Now()
	data.ReceiveCode = form.ReceiveCode

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
	opts := options.Find().SetSort(bson.D{{"sequence", 1}})
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"productId": product, "unitId": unit}, opts)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := entities.ProductStock{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductStock{}
	}
	return items, nil
}

func (entity *productEntity) GetProductStockMaxSequence(productId string, unitId string) int {
	logrus.Info("GetProductStockMaxSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)
	opts := options.FindOne().SetSort(bson.D{{"sequence", -1}})
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
	objId, _ := primitive.ObjectIDFromHex(stockId)
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Quantity = data.Quantity + quantity

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	logrus.Info("RemoveProductStockQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(stockId)
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Quantity = data.Quantity - quantity

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
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
	for cursor.Next(ctx) {
		item := entities.ProductUnit{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductUnit{}
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
	for cursor.Next(ctx) {
		item := entities.ProductPrice{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductPrice{}
	}
	return items, nil
}

func (entity *productEntity) CreateProductPrice(form request.ProductPrice) (*entities.ProductPrice, error) {
	logrus.Info("CreateProductPrice")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductPrice{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(form.ProductId)
	data.UnitId, _ = primitive.ObjectIDFromHex(form.UnitId)
	data.CustomerType = form.CustomerType
	data.Price = form.Price
	_, err := entity.productPricesRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductPriceById(id string, form request.ProductPrice) (*entities.ProductPrice, error) {
	logrus.Info("UpdateProductPriceById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductPrice{}
	err := entity.productPricesRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.CustomerType = form.CustomerType
	data.Price = form.Price
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productPricesRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductPriceById(id string) (*entities.ProductPrice, error) {
	logrus.Info("RemoveProductPriceById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductPrice{}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.productPricesRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	if data.CustomerType == constant.CustomerTypeGeneral {
		return nil, errors.New("can not remove default price")
	}
	_, err = entity.productPricesRepo.DeleteOne(ctx, bson.M{"_id": objId})
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

func (entity *productEntity) CreateProductUnit(form request.ProductUnit) (*entities.ProductUnit, error) {
	logrus.Info("CreateProductUnit")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductUnit{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(form.ProductId)
	data.Unit = form.Unit
	data.Size = form.Size
	data.CostPrice = form.CostPrice
	data.Volume = form.Volume
	data.VolumeUnit = form.VolumeUnit
	data.Barcode = form.Barcode
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

func (entity *productEntity) UpdateProductUnitById(id string, form request.ProductUnit) (*entities.ProductUnit, error) {
	logrus.Info("UpdateProductUnitById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductUnit{}
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Unit = form.Unit
	data.Size = form.Size
	data.CostPrice = form.CostPrice
	data.Volume = form.Volume
	data.VolumeUnit = form.VolumeUnit
	data.Barcode = form.Barcode
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productUnitsRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
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
	_, err = entity.productUnitsRepo.DeleteOne(ctx, bson.M{"_id": objId})
	return &data, nil
}

func (entity *productEntity) CreateProductStock(form request.ProductStock) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStock")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductStock{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(form.ProductId)
	data.UnitId, _ = primitive.ObjectIDFromHex(form.UnitId)
	data.Sequence = entity.GetProductStockMaxSequence(form.ProductId, form.UnitId) + 1
	data.LotNumber = form.LotNumber
	data.CostPrice = form.CostPrice
	data.Price = form.Price
	data.Import = form.Quantity
	data.Quantity = form.Quantity
	data.ExpireDate = form.ExpireDate
	data.ImportDate = form.ImportDate
	data.ReceiveCode = form.ReceiveCode
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
	opts := options.Find().SetSort(bson.D{{"sequence", 1}})
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"productId": product}, opts)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := entities.ProductStock{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []entities.ProductStock{}
	}
	return items, nil

}

func (entity *productEntity) UpdateProductStockById(id string, form request.UpdateProductStock) (*entities.ProductStock, error) {
	logrus.Info("UpdateProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.LotNumber = form.LotNumber
	data.CostPrice = form.CostPrice
	data.Price = form.Price
	data.ExpireDate = form.ExpireDate
	data.ImportDate = form.ImportDate
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error) {
	logrus.Info("UpdateProductStockQuantityById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := entities.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	data.Quantity = quantity
	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductStockById(id string) (*entities.ProductStock, error) {
	logrus.Info("RemoveProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductStock{}
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.productStockRepo.DeleteOne(ctx, bson.M{"_id": objId})
	return &data, nil
}

func (entity *productEntity) UpdateProductStockSequence(form request.UpdateProductStockSequence) ([]entities.ProductStock, error) {
	logrus.Info("UpdateProductStockSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objIds := make([]primitive.ObjectID, 0, len(form.Stocks))
	for _, value := range form.Stocks {
		id, err := primitive.ObjectIDFromHex(value.StockId)
		if err != nil {
			return nil, err
		}
		objIds = append(objIds, id)
	}
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"_id": bson.M{"$in": objIds}})
	if err != nil {
		return nil, err
	}
	stocks := make([]entities.ProductStock, 0, len(form.Stocks))
	for cursor.Next(ctx) {
		item := entities.ProductStock{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			stocks = append(stocks, item)
		}
	}
	if stocks == nil {
		stocks = []entities.ProductStock{}
	}
	for i, value := range stocks {
		value.Sequence = form.Stocks[i].Sequence
		_, err = entity.productStockRepo.ReplaceOne(ctx, bson.M{"_id": value.Id}, value)
		if err != nil {
			logrus.Error(err)
		}
	}
	return stocks, nil
}

func (entity *productEntity) CreateProductHistory(form request.ProductHistory) (*entities.ProductHistory, error) {
	logrus.Info("CreateProductHistory")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := entities.ProductHistory{}
	data.Id = primitive.NewObjectID()
	data.ProductId, _ = primitive.ObjectIDFromHex(form.ProductId)
	data.Description = form.Description
	data.Type = form.Type
	data.Unit = form.Unit
	data.CostPrice = form.CostPrice
	data.Price = form.Price
	data.Quantity = form.Quantity
	data.Import = form.Import
	data.CreatedBy = form.CreatedBy
	data.CreatedDate = time.Now()
	data.Balance = form.Balance

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
	cursor, err := entity.productStockRepo.Find(ctx, bson.M{"productId": product, "unitId": unit})
	balance := 0
	if err != nil {
		return balance
	}
	for cursor.Next(ctx) {
		item := entities.ProductStock{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			balance += item.Quantity
		}
	}
	return balance
}
