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
	"pos/app/domain/model"
	"pos/app/domain/request"
	"pos/db"
	"strings"
	"time"
)

type productEntity struct {
	productsRepo      *mongo.Collection
	productPricesRepo *mongo.Collection
	productLotsRepo   *mongo.Collection
	productUnitsRepo  *mongo.Collection
	productStockRepo  *mongo.Collection
}

type IProduct interface {
	CreateIndex()
	GetProductAll(product request.GetProduct) ([]model.ProductDetail, error)
	GetProductBySerialNumber(serialNumber string) (*model.Product, error)
	GetProductById(id string) (*model.Product, error)
	CreateProduct(form request.Product) (*model.Product, error)
	RemoveProductById(id string) (*model.Product, error)
	UpdateProductById(id string, form request.UpdateProduct) (*model.Product, error)
	RemoveQuantityById(id string, quantity int) (*model.Product, error)
	AddQuantityById(id string, quantity int) (*model.Product, error)
	GetTotalCostPrice(id string, quantity int) float64

	CreateProductLotByProductId(productId string, form request.Product) (*model.ProductLot, error)
	CreateProductLot(form request.ProductLot) (*model.ProductLot, error)
	GetProductLots(form request.GetExpireRange) ([]model.ProductLot, error)
	GetProductLotsByProductId(productId string) ([]model.ProductLot, error)
	GetProductLotsByIds(ids []string) ([]model.ProductLot, error)
	GetProductLotsExpired() ([]model.ProductLot, error)
	GetProductLotsExpireNotify(form request.GetExpireRange) ([]model.ProductLotDetail, error)
	GetProductLotById(id string) (*model.ProductLot, error)
	RemoveProductLotById(id string) (*model.ProductLot, error)
	UpdateProductLotById(id string, form request.UpdateProductLot) (*model.ProductLot, error)
	UpdateProductLotNotifyById(id string, form request.UpdateProductLotNotify) (*model.ProductLot, error)
	UpdateProductLotQuantityById(id string, form request.UpdateProductLotQuantity) (*model.ProductLot, error)

	CreateProductUnitByProductId(productId string, form request.Product) (*model.ProductUnit, error)
	GetProductUnitByProductIdAndUnit(productId string, unit string) (*model.ProductUnit, error)
	GetProductUnitsByProductId(productId string) ([]model.ProductUnit, error)

	CreateProductPriceByProductAndUnitId(productId string, unitId string, form request.Product) (*model.ProductPrice, error)
	GetProductPricesByProductId(productId string) ([]model.ProductPrice, error)

	CreateProductStockByProductAndUnitId(productId string, unitId string, form request.Product) (*model.ProductStock, error)
	GetProductStocksByProductAndUnitId(productId string, unitId string) ([]model.ProductStock, error)
	GetProductStockMaxSequence(productId string, unitId string) int
	RemoveProductStockQuantityByProductAndUnitId(productId string, unitId string, quantity int) (*model.ProductStock, error)
	AddProductStockQuantityByProductAndUnitId(productId string, unitId string, quantity int) (*model.ProductStock, error)
}

func NewProductEntity(resource *db.Resource) IProduct {
	productsRepo := resource.PosDb.Collection("products")
	productPricesRepo := resource.PosDb.Collection("product_prices")
	productUnitsRepo := resource.PosDb.Collection("product_units")
	productLotsRepo := resource.PosDb.Collection("product_lots")
	productStockRepo := resource.PosDb.Collection("product_stocks")
	entity := &productEntity{
		productsRepo:      productsRepo,
		productPricesRepo: productPricesRepo,
		productLotsRepo:   productLotsRepo,
		productUnitsRepo:  productUnitsRepo,
		productStockRepo:  productStockRepo,
	}
	entity.CreateIndex()
	return entity
}

func (entity *productEntity) CreateIndex() {
	ctx, cancel := utils.InitContext()
	defer cancel()
	mod := mongo.IndexModel{
		Keys: bson.M{
			"serialNumber": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := entity.productsRepo.Indexes().CreateOne(ctx, mod)
	if err != nil {
		logrus.Error(err)
	}
}

func (entity *productEntity) GetProductAll(product request.GetProduct) (items []model.ProductDetail, err error) {
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
	}
	cursor, err := entity.productsRepo.Aggregate(ctx, pipeline)

	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := model.ProductDetail{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductDetail{}
	}
	return items, nil
}

func (entity *productEntity) GetProductBySerialNumber(serialNumber string) (*model.Product, error) {
	logrus.Info("GetProductBySerialNumber")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Product
	err := entity.productsRepo.FindOne(ctx, bson.M{"serialNumber": serialNumber}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProduct(form request.Product) (*model.Product, error) {
	logrus.Info("CreateProduct")
	ctx, cancel := utils.InitContext()
	defer cancel()
	serialNumber := strings.TrimSpace(form.SerialNumber)
	data, _ := entity.GetProductBySerialNumber(serialNumber)
	if data != nil {
		data.Name = form.Name
		data.NameEn = form.NameEn
		data.Description = form.Description
		data.SerialNumber = serialNumber
		data.Price = form.Price
		data.CostPrice = form.CostPrice
		data.Unit = form.Unit
		data.Quantity = data.Quantity + form.Quantity
		data.Category = form.Category
		data.UpdatedDate = time.Now()

		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err := entity.productsRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		data := model.Product{}
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
}

func (entity *productEntity) GetProductById(id string) (*model.Product, error) {
	logrus.Info("GetProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Product
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.productsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) RemoveProductById(id string) (*model.Product, error) {
	logrus.Info("RemoveProductById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.Product
	objId, _ := primitive.ObjectIDFromHex(id)
	err := entity.productsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.productsRepo.DeleteOne(ctx, bson.M{"_id": objId})
	return &data, nil
}

func (entity *productEntity) UpdateProductById(id string, form request.UpdateProduct) (*model.Product, error) {
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

func (entity *productEntity) RemoveQuantityById(id string, quantity int) (*model.Product, error) {
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

func (entity *productEntity) AddQuantityById(id string, quantity int) (*model.Product, error) {
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

func (entity *productEntity) CreateProductLotByProductId(productId string, form request.Product) (*model.ProductLot, error) {
	logrus.Info("CreateProductLotByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ProductLot{}
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

func (entity *productEntity) CreateProductLot(form request.ProductLot) (*model.ProductLot, error) {
	logrus.Info("CreateProductLot")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ProductLot{}
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

func (entity *productEntity) GetProductLots(form request.GetExpireRange) (items []model.ProductLot, err error) {
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
		data := model.ProductLot{}
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []model.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsByProductId(productId string) (items []model.ProductLot, err error) {
	logrus.Info("GetProductLotsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productLotsRepo.Find(ctx, bson.M{"productId": objId})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var productLot model.ProductLot
		err = cursor.Decode(&productLot)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, productLot)
		}
	}
	if items == nil {
		items = []model.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) RemoveProductLotById(id string) (*model.ProductLot, error) {
	logrus.Info("RemoveProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ProductLot{}
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

func (entity *productEntity) GetProductLotsByIds(ids []string) (items []model.ProductLot, err error) {
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
		item := model.ProductLot{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpired() (items []model.ProductLot, err error) {
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
		item := model.ProductLot{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductLot{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotsExpireNotify(form request.GetExpireRange) (items []model.ProductLotDetail, err error) {
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
		data := model.ProductLotDetail{}
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []model.ProductLotDetail{}
	}
	return items, nil
}

func (entity *productEntity) GetProductLotById(id string) (*model.ProductLot, error) {
	logrus.Info("GetProductLotById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, _ := primitive.ObjectIDFromHex(id)
	data := model.ProductLot{}
	err := entity.productLotsRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) UpdateProductLotById(id string, form request.UpdateProductLot) (*model.ProductLot, error) {
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

func (entity *productEntity) UpdateProductLotNotifyById(id string, form request.UpdateProductLotNotify) (*model.ProductLot, error) {
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

func (entity *productEntity) UpdateProductLotQuantityById(id string, form request.UpdateProductLotQuantity) (*model.ProductLot, error) {
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

func (entity *productEntity) CreateProductUnitByProductId(productId string, form request.Product) (*model.ProductUnit, error) {
	logrus.Info("CreateProductUnitByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ProductUnit{}
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

func (entity *productEntity) GetProductUnitByProductIdAndUnit(productId string, unit string) (*model.ProductUnit, error) {
	logrus.Info("GetProductUnitByProductIdAndUnit")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data := model.ProductUnit{}
	product, _ := primitive.ObjectIDFromHex(productId)
	err := entity.productUnitsRepo.FindOne(ctx, bson.M{"productId": product, "unit": unit}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductPriceByProductAndUnitId(productId string, unitId string, form request.Product) (*model.ProductPrice, error) {
	logrus.Info("CreateProductPriceByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := model.ProductPrice{}
	product, _ := primitive.ObjectIDFromHex(productId)
	unit, _ := primitive.ObjectIDFromHex(unitId)

	err := entity.productPricesRepo.FindOne(ctx, bson.M{"productId": product, "unitId": unit}).Decode(&data)
	if err != nil {
		data.Id = primitive.NewObjectID()
		data.ProductId, _ = primitive.ObjectIDFromHex(productId)
		data.UnitId, _ = primitive.ObjectIDFromHex(unitId)
		data.CustomerType = constant.CustomerTypeGeneral
		data.Price = form.Price
		_, err = entity.productPricesRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
	} else {
		data.Price = form.Price
		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err = entity.productPricesRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
	}

	return &data, nil
}

func (entity *productEntity) CreateProductStockByProductAndUnitId(productId string, unitId string, form request.Product) (*model.ProductStock, error) {
	logrus.Info("CreateProductStockByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()

	data := model.ProductStock{}
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
	_, err := entity.productStockRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductStocksByProductAndUnitId(productId string, unitId string) (items []model.ProductStock, err error) {
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
		item := model.ProductStock{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductStock{}
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
	data := model.ProductStock{}
	err := entity.productStockRepo.FindOne(ctx, bson.M{"productId": product, "unitId": unit}, opts).Decode(&data)
	if err != nil {
		return 0
	}
	return data.Sequence
}

func (entity *productEntity) RemoveProductStockQuantityByProductAndUnitId(productId string, unitId string, quantity int) (*model.ProductStock, error) {
	logrus.Info("RemoveProductStockQuantityByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	stocks, err := entity.GetProductStocksByProductAndUnitId(productId, unitId)
	if err != nil {
		return nil, err
	}
	if len(stocks) == 0 {
		return nil, errors.New("not found product stock")
	}

	data := stocks[0]
	for _, stock := range stocks {
		if stock.Quantity > 0 {
			data = stock
			break
		}
	}

	data.Quantity = data.Quantity - quantity
	isReturnNewDoc := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, updateOpts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) AddProductStockQuantityByProductAndUnitId(productId string, unitId string, quantity int) (*model.ProductStock, error) {
	logrus.Info("AddProductStockQuantityByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	stocks, err := entity.GetProductStocksByProductAndUnitId(productId, unitId)
	if err != nil {
		return nil, err
	}
	if len(stocks) == 0 {
		return nil, errors.New("not found product stock")
	}

	data := stocks[0]
	for _, stock := range stocks {
		if stock.Quantity > 0 {
			data = stock
			break
		}
	}

	data.Quantity = data.Quantity + quantity
	isReturnNewDoc := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productStockRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, updateOpts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) GetProductUnitsByProductId(productId string) (items []model.ProductUnit, err error) {
	logrus.Info("GetProductUnitsByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productUnitsRepo.Find(ctx, bson.M{"productId": product})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := model.ProductUnit{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductUnit{}
	}
	return items, nil
}

func (entity *productEntity) GetProductPricesByProductId(productId string) (items []model.ProductPrice, err error) {
	logrus.Info("GetProductPricesByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, _ := primitive.ObjectIDFromHex(productId)
	cursor, err := entity.productPricesRepo.Find(ctx, bson.M{"productId": product})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		item := model.ProductPrice{}
		err = cursor.Decode(&item)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []model.ProductPrice{}
	}
	return items, nil
}
