package repository

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pos/app/core/utils"
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
}

type IProduct interface {
	CreateIndex() (string, error)
	GetProductAll(product request.GetProduct) ([]model.Product, error)
	GetProductBySerialNumber(serialNumber string) (*model.Product, error)
	GetProductById(id string) (*model.Product, error)
	CreateProduct(form request.Product) (*model.Product, error)
	RemoveProductById(id string) (*model.Product, error)
	UpdateProductById(id string, form request.UpdateProduct) (*model.Product, error)
	RemoveQuantityById(id string, quantity int) (*model.Product, error)
	AddQuantityById(id string, quantity int) (*model.Product, error)
	GetTotalCostPrice(id string, quantity int) float64

	GetProductPriceByProductCustomerId(productId string, customerId string) (*model.ProductPrice, error)
	CreateProductPriceByProductId(form request.ProductPrice) (*model.ProductPrice, error)
	GetProductPriceDetailByProductId(productId string) ([]model.ProductPriceDetail, error)
	GetProductPriceDetailByCustomerId(customerId string) ([]model.ProductPriceDetail, error)

	CreateProductLot(productId string, form request.Product) (*model.ProductLot, error)
	GetProductLotsByProductId(productId string) ([]model.ProductLot, error)
	GetProductLotsExpire(form request.GetExpireRange) ([]model.ProductLotDetail, error)
	GetProductLotById(id string) (*model.ProductLot, error)
	UpdateProductLotById(id string, form request.ProductLot) (*model.ProductLot, error)
}

func NewProductEntity(resource *db.Resource) IProduct {
	productsRepo := resource.PosDb.Collection("products")
	productPricesRepo := resource.PosDb.Collection("product_prices")
	productLotsRepo := resource.PosDb.Collection("product_lots")
	var entity IProduct = &productEntity{
		productsRepo:      productsRepo,
		productPricesRepo: productPricesRepo,
		productLotsRepo:   productLotsRepo,
	}
	_, _ = entity.CreateIndex()
	return entity
}

func (entity *productEntity) CreateIndex() (string, error) {
	ctx, cancel := utils.InitContext()
	defer cancel()
	mod := mongo.IndexModel{
		Keys: bson.M{
			"serialNumber": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	ind, err := entity.productsRepo.Indexes().CreateOne(ctx, mod)
	return ind, err
}

func (entity *productEntity) GetProductAll(product request.GetProduct) ([]model.Product, error) {
	logrus.Info("GetProductAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	query := bson.M{}
	if product.Category != "" {
		query["category"] = product.Category
	}
	cursor, err := entity.productsRepo.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	var products []model.Product
	for cursor.Next(ctx) {
		var user model.Product
		err = cursor.Decode(&user)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			products = append(products, user)
		}
	}
	if products == nil {
		products = []model.Product{}
	}
	return products, nil
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

func (entity *productEntity) GetProductPriceByProductCustomerId(productId string, customerId string) (*model.ProductPrice, error) {
	logrus.Info("GetProductPriceByProductCustomerId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var data model.ProductPrice
	pid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cid, err := primitive.ObjectIDFromHex(customerId)
	if err != nil {
		return nil, err
	}
	err = entity.productPricesRepo.FindOne(ctx, bson.M{"productId": pid, "customerId": cid}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductPriceByProductId(form request.ProductPrice) (data *model.ProductPrice, err error) {
	logrus.Info("CreateProductPriceByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data, _ = entity.GetProductPriceByProductCustomerId(form.ProductId, form.CustomerId)
	if data != nil {
		data.CustomerPrice = form.CustomerPrice
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()

		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err = entity.productPricesRepo.FindOneAndUpdate(ctx, bson.M{"productId": data.ProductId, "customerId": data.CustomerId}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		pid, err := primitive.ObjectIDFromHex(form.ProductId)
		if err != nil {
			return nil, err
		}
		data.ProductId = pid
		cid, err := primitive.ObjectIDFromHex(form.CustomerId)
		if err != nil {
			return nil, err
		}
		data.CustomerId = cid
		data.CustomerPrice = form.CustomerPrice
		data.CreatedBy = form.CreatedBy
		data.CreatedDate = time.Now()
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()
		_, err = entity.productPricesRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func (entity *productEntity) GetProductPriceDetailByProductId(productId string) ([]model.ProductPriceDetail, error) {
	logrus.Info("GetProductPriceDetailByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	pid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.productPricesRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"productId": pid,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "customers",
				"localField":   "customerId",
				"foreignField": "_id",
				"as":           "customer",
			},
		},
		{"$unwind": "customer"},
	})
	if err != nil {
		return nil, err
	}
	var items []model.ProductPriceDetail
	for cursor.Next(ctx) {
		var data model.ProductPriceDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []model.ProductPriceDetail{}
	}
	return items, nil
}

func (entity *productEntity) GetProductPriceDetailByCustomerId(customerId string) ([]model.ProductPriceDetail, error) {
	logrus.Info("GetProductPriceDetailByCustomerId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cid, err := primitive.ObjectIDFromHex(customerId)
	if err != nil {
		return nil, err
	}
	cursor, err := entity.productPricesRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"customerId": cid,
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
		{"$unwind": "product"},
	})
	if err != nil {
		return nil, err
	}
	var items []model.ProductPriceDetail
	for cursor.Next(ctx) {
		var data model.ProductPriceDetail
		err = cursor.Decode(&data)
		if err != nil {
			logrus.Error(err)
			logrus.Info(cursor.Current)
		} else {
			items = append(items, data)
		}
	}
	if items == nil {
		items = []model.ProductPriceDetail{}
	}
	return items, nil
}

func (entity *productEntity) CreateProductLot(productId string, form request.Product) (*model.ProductLot, error) {
	logrus.Info("CreateProductLot")
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
	data.UpdatedBy = form.CreatedBy
	data.CreatedDate = time.Now()
	data.UpdatedDate = time.Now()

	_, err := entity.productLotsRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
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

func (entity *productEntity) GetProductLotsExpire(form request.GetExpireRange) (items []model.ProductLotDetail, err error) {
	logrus.Info("GetProductLotsExpire")
	ctx, cancel := utils.InitContext()
	defer cancel()
	cursor, err := entity.productLotsRepo.Aggregate(ctx, []bson.M{
		{
			"$match": bson.M{
				"expireDate": bson.M{
					"$gt": form.StartDate,
					"$lt": form.EndDate,
				},
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
		var data model.ProductLotDetail
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

func (entity *productEntity) UpdateProductLotById(id string, form request.ProductLot) (*model.ProductLot, error) {
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
