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
	productRepo      *mongo.Collection
	productPriceRepo *mongo.Collection
}

type IProduct interface {
	CreateIndex() (string, error)
	GetProductAll() ([]model.Product, error)
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
}

func NewProductEntity(resource *db.Resource) IProduct {
	productRepo := resource.PosDb.Collection("products")
	productPriceRepo := resource.PosDb.Collection("products_price")
	var entity IProduct = &productEntity{
		productRepo:      productRepo,
		productPriceRepo: productPriceRepo,
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
	ind, err := entity.productRepo.Indexes().CreateOne(ctx, mod)
	return ind, err
}

func (entity *productEntity) GetProductAll() ([]model.Product, error) {
	logrus.Info("GetProductAll")
	ctx, cancel := utils.InitContext()
	defer cancel()
	var products []model.Product
	cursor, err := entity.productRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
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
	err := entity.productRepo.FindOne(ctx, bson.M{"serialNumber": serialNumber}).Decode(&data)
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
		data.UpdatedDate = time.Now()

		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err := entity.productRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
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
		data.CreatedDate = time.Now()
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()
		_, err := entity.productRepo.InsertOne(ctx, data)
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
	err := entity.productRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
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
	err := entity.productRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	_, err = entity.productRepo.DeleteOne(ctx, bson.M{"_id": objId})
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
	data.UpdatedBy = form.UpdatedBy
	data.UpdatedDate = time.Now()

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.productRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
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
	err = entity.productRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
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
	err = entity.productRepo.FindOneAndUpdate(ctx, bson.M{"_id": objId}, bson.M{"$set": data}, opts).Decode(&data)
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
	err = entity.productPriceRepo.FindOne(ctx, bson.M{"productId": pid, "customerId": cid}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productEntity) CreateProductPriceByProductId(form request.ProductPrice) (*model.ProductPrice, error) {
	logrus.Info("CreateProductPriceByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	data, _ := entity.GetProductPriceByProductCustomerId(form.ProductId, form.CustomerId)
	if data != nil {
		data.CustomerPrice = form.CustomerPrice
		data.UpdatedBy = form.CreatedBy
		data.UpdatedDate = time.Now()

		isReturnNewDoc := options.After
		opts := &options.FindOneAndUpdateOptions{
			ReturnDocument: &isReturnNewDoc,
		}
		err := entity.productPriceRepo.FindOneAndUpdate(ctx, bson.M{"_id": data.Id}, bson.M{"$set": data}, opts).Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		data := model.ProductPrice{}
		data.Id = primitive.NewObjectID()
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
		_, err = entity.productPriceRepo.InsertOne(ctx, data)
		if err != nil {
			return nil, err
		}
		return &data, nil
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
	cursor, err := entity.productPriceRepo.Aggregate(ctx, []bson.M{
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
	cursor, err := entity.productPriceRepo.Aggregate(ctx, []bson.M{
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
