package repositories

import (
	"context"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productStockEntity struct {
	productStockRepo   *mongo.Collection
	productHistoryRepo *mongo.Collection
}

type IProductStock interface {
	// ProductStock
	CreateProductStock(param request.ProductStock) (*entities.ProductStock, error)
	GetProductStockById(id string) (*entities.ProductStock, error)
	UpdateProductStockById(id string, param request.UpdateProductStock) (*entities.ProductStock, error)
	UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error)
	UpdateProductStockSequence(param request.UpdateProductStockSequence) ([]entities.ProductStock, error)
	RemoveProductStockById(id string) (*entities.ProductStock, error)
	GetProductStocksByProductId(productId string, branchId string) ([]entities.ProductStock, error)
	GetProductStocksByProductAndUnitId(productId string, unitId string, branchId string) ([]entities.ProductStock, error)
	GetProductStockMaxSequence(productId string, unitId string, branchId string) int
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
	GetExpiringProductStocks(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error)
}

func newProductStockEntity(resource *db.Resource) *productStockEntity {
	productStockRepo := resource.PosDb.Collection("product_stocks")
	productHistoryRepo := resource.PosDb.Collection("product_histories")
	entity := &productStockEntity{
		productStockRepo:   productStockRepo,
		productHistoryRepo: productHistoryRepo,
	}
	ensureProductStockIndexes(productStockRepo, productHistoryRepo)
	return entity
}

func NewProductStockEntity(resource *db.Resource) IProductStock {
	return newProductStockEntity(resource)
}

func ensureProductStockIndexes(productStockRepo *mongo.Collection, productHistoryRepo *mongo.Collection) {
	createCollectionIndex(productStockRepo, "product_stocks branchId+productId", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "productId", Value: 1}},
	})
	createCollectionIndex(productHistoryRepo, "product_histories branchId+createdDate", mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}, {Key: "createdDate", Value: -1}},
	})
	createCollectionIndex(productHistoryRepo, "product_histories productId", mongo.IndexModel{
		Keys: bson.D{{Key: "productId", Value: 1}},
	})
}

// --- ProductStock CRUD ---

func (entity *productStockEntity) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	logrus.Info("CreateProductStock")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductStockWithContext(ctx, param)
}

func (entity *productStockEntity) createProductStockWithContext(ctx context.Context, param request.ProductStock) (*entities.ProductStock, error) {
	branchObjID, err := primitive.ObjectIDFromHex(param.BranchId)
	if err != nil {
		return nil, err
	}
	productObjID, err := primitive.ObjectIDFromHex(param.ProductId)
	if err != nil {
		return nil, err
	}
	unitObjID, err := primitive.ObjectIDFromHex(param.UnitId)
	if err != nil {
		return nil, err
	}
	data := entities.ProductStock{}
	data.Id = primitive.NewObjectID()
	data.BranchId = branchObjID
	data.ProductId = productObjID
	data.UnitId = unitObjID
	data.Sequence = entity.getProductStockMaxSequenceWithContext(ctx, param.ProductId, param.UnitId, param.BranchId) + 1
	data.LotNumber = param.LotNumber
	data.CostPrice = param.CostPrice
	data.Price = param.Price
	data.Import = param.Quantity
	data.Quantity = param.Quantity
	data.ExpireDate = param.ExpireDate.Time
	data.ImportDate = param.ImportDate.Time
	data.ReceiveCode = param.ReceiveCode
	_, err = entity.productStockRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productStockEntity) GetProductStockById(id string) (*entities.ProductStock, error) {
	logrus.Info("GetProductStockById")
	ctx, cancel := utils.InitContext()
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data := entities.ProductStock{}
	err = entity.productStockRepo.FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productStockEntity) GetProductStocksByProductId(productId string, branchId string) (items []entities.ProductStock, err error) {
	logrus.Info("GetProductStocksByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"productId": product}
	if branchId != "" {
		branch, branchErr := primitive.ObjectIDFromHex(branchId)
		if branchErr != nil {
			return nil, branchErr
		}
		filter["branchId"] = branch
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

func (entity *productStockEntity) GetProductStocksByProductAndUnitId(productId string, unitId string, branchId string) (items []entities.ProductStock, err error) {
	logrus.Info("GetProductStocksByProductAndUnitId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}})
	filter, err := buildProductStockSequenceFilter(productId, unitId, branchId)
	if err != nil {
		return nil, err
	}
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

func (entity *productStockEntity) UpdateProductStockById(id string, param request.UpdateProductStock) (*entities.ProductStock, error) {
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
		"expireDate": param.ExpireDate.Time,
		"importDate": param.ImportDate.Time,
	}}, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productStockEntity) UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error) {
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

func (entity *productStockEntity) RemoveProductStockById(id string) (*entities.ProductStock, error) {
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

func (entity *productStockEntity) UpdateProductStockSequence(param request.UpdateProductStockSequence) ([]entities.ProductStock, error) {
	logrus.Info("UpdateProductStockSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()

	branchFilter, err := buildProductStockSequenceBranchFilter(param.BranchId)
	if err != nil {
		return nil, err
	}
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
		filter := bson.M{"_id": id}
		for k, v := range branchFilter {
			filter[k] = v
		}
		writes = append(writes, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(bson.M{"$set": bson.M{"sequence": sequenceMap[id.Hex()]}}))
	}
	if _, err := entity.productStockRepo.BulkWrite(ctx, writes); err != nil {
		return nil, err
	}

	findFilter := bson.M{"_id": bson.M{"$in": objIds}}
	for k, v := range branchFilter {
		findFilter[k] = v
	}
	cursor, err := entity.productStockRepo.Find(ctx, findFilter,
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

func buildProductStockSequenceBranchFilter(branchId string) (bson.M, error) {
	if branchId == "" {
		return bson.M{}, nil
	}
	branchObjID, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}
	return bson.M{"branchId": branchObjID}, nil
}

// --- Stock queries ---

func (entity *productStockEntity) GetProductStockMaxSequence(productId string, unitId string, branchId string) int {
	logrus.Info("GetProductStockMaxSequence")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.getProductStockMaxSequenceWithContext(ctx, productId, unitId, branchId)
}

func (entity *productStockEntity) getProductStockMaxSequenceWithContext(ctx context.Context, productId string, unitId string, branchId string) int {
	filter, err := buildProductStockSequenceFilter(productId, unitId, branchId)
	if err != nil {
		return 0
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "sequence", Value: -1}})
	data := entities.ProductStock{}
	err = entity.productStockRepo.FindOne(ctx, filter, opts).Decode(&data)
	if err != nil {
		return 0
	}
	return data.Sequence
}

func buildProductStockSequenceFilter(productId string, unitId string, branchId string) (bson.M, error) {
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	unit, err := primitive.ObjectIDFromHex(unitId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"productId": product, "unitId": unit}
	if branchId != "" {
		branch, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		filter["branchId"] = branch
	}
	return filter, nil
}

func (entity *productStockEntity) GetProductStockBalance(productId string, unitId string, branchId string) int {
	logrus.Info("GetProductStockBalance")
	ctx, cancel := utils.InitContext()
	defer cancel()
	balance, err := entity.getProductStockBalanceWithContext(ctx, productId, unitId, branchId)
	if err != nil {
		return 0
	}
	return balance
}

func (entity *productStockEntity) getProductStockBalanceWithContext(ctx context.Context, productId string, unitId string, branchId string) (int, error) {
	product, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return 0, err
	}
	unit, err := primitive.ObjectIDFromHex(unitId)
	if err != nil {
		return 0, err
	}
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

// --- Stock quantity adjustments ---

func (entity *productStockEntity) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
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

func (entity *productStockEntity) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
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

// --- ProductHistory ---

func (entity *productStockEntity) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	logrus.Info("CreateProductHistory")
	ctx, cancel := utils.InitContext()
	defer cancel()
	return entity.createProductHistoryWithContext(ctx, param)
}

func (entity *productStockEntity) createProductHistoryWithContext(ctx context.Context, param request.ProductHistory) (*entities.ProductHistory, error) {
	branchObjID, err := primitive.ObjectIDFromHex(param.BranchId)
	if err != nil {
		return nil, err
	}
	productObjID, err := primitive.ObjectIDFromHex(param.ProductId)
	if err != nil {
		return nil, err
	}
	data := entities.ProductHistory{}
	data.Id = primitive.NewObjectID()
	data.BranchId = branchObjID
	data.ProductId = productObjID
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

	_, err = entity.productHistoryRepo.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (entity *productStockEntity) RemoveProductHistoryById(id string) (*entities.ProductHistory, error) {
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

func (entity *productStockEntity) GetProductHistoryByProductId(productId string, branchId string) ([]entities.ProductHistory, error) {
	logrus.Info("GetProductHistoryByProductId")
	ctx, cancel := utils.InitContext()
	defer cancel()
	prodObjId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, err
	}
	filter, err := buildProductHistoryFilter(prodObjId, branchId)
	if err != nil {
		return nil, err
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

func (entity *productStockEntity) GetProductHistoryByDateRange(branchId string, startDate time.Time, endDate time.Time) ([]entities.ProductHistory, error) {
	logrus.Info("GetProductHistoryByDateRange")
	ctx, cancel := utils.InitContext()
	defer cancel()
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	filter := bson.M{
		"createdDate": bson.M{"$gte": startDate, "$lte": endOfDay},
	}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
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

func buildProductHistoryFilter(productId primitive.ObjectID, branchId string) (bson.M, error) {
	filter := bson.M{"productId": productId}
	if branchId == "" {
		return filter, nil
	}
	branchObjId, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}
	filter["branchId"] = branchObjId
	return filter, nil
}

// --- Reports ---

func (entity *productStockEntity) GetExpiringProductStocks(param request.GetProductLotsExpireRange, branchId string) (items []entities.ProductLotDetail, err error) {
	logrus.Info("GetExpiringProductStocks")
	ctx, cancel := utils.InitContext()
	defer cancel()

	pipeline, err := buildExpiringProductStocksPipeline(param, branchId)
	if err != nil {
		return nil, err
	}

	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	items = []entities.ProductLotDetail{}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func buildExpiringProductStocksPipeline(param request.GetProductLotsExpireRange, branchId string) ([]bson.M, error) {
	matchFilter := bson.M{
		"expireDate": bson.M{
			"$gte": param.StartDate.Time,
			"$lt":  param.EndDate.Time,
		},
		"quantity": bson.M{"$gt": 0},
	}
	if branchId != "" {
		branchObjID, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjID
	}

	return []bson.M{
		{"$match": matchFilter},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "productId",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": "$product"},
		{"$match": bson.M{"product.deletedDate": bson.M{"$exists": false}}},
		{"$project": bson.M{
			"_id":         "$_id",
			"productId":   "$productId",
			"lotNumber":   "$lotNumber",
			"costPrice":   "$costPrice",
			"quantity":    "$quantity",
			"expireDate":  "$expireDate",
			"createdDate": "$importDate",
			"updatedDate": "$importDate",
			"notify":      true,
			"product":     "$product",
		}},
		{"$sort": bson.M{"expireDate": 1, "lotNumber": 1}},
	}, nil
}

func (entity *productStockEntity) GetLowStockProducts(threshold int, branchId string) ([]entities.LowStockProduct, error) {
	logrus.Info("GetLowStockProducts")
	ctx, cancel := utils.InitContext()
	defer cancel()
	pipeline, err := buildLowStockProductsPipeline(threshold, branchId)
	if err != nil {
		return nil, err
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

func buildLowStockProductsPipeline(threshold int, branchId string) ([]bson.M, error) {
	matchStage := bson.M{}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchStage["branchId"] = branchObjId
	}

	return []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id": bson.M{
				"productId": "$productId",
				"unitId":    "$unitId",
			},
			"totalStock": bson.M{"$sum": "$quantity"},
		}},
		{"$match": bson.M{"totalStock": bson.M{"$lte": threshold}}},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "_id.productId",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": "$product"},
		{"$lookup": bson.M{
			"from":         "product_units",
			"localField":   "_id.unitId",
			"foreignField": "_id",
			"as":           "unit",
		}},
		{"$unwind": "$unit"},
		{"$project": bson.M{
			"_id":          "$_id.productId",
			"totalStock":   1,
			"name":         "$product.name",
			"serialNumber": "$product.serialNumber",
			"unit":         "$unit.unit",
		}},
		{"$sort": bson.M{"totalStock": 1, "name": 1, "unit": 1}},
	}, nil
}

func (entity *productStockEntity) GetStockReport(branchId string) ([]entities.StockReport, error) {
	logrus.Info("GetStockReport")
	ctx, cancel := utils.InitContext()
	defer cancel()
	pipeline, err := buildStockReportPipeline(branchId)
	if err != nil {
		return nil, err
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

func buildStockReportPipeline(branchId string) ([]bson.M, error) {
	matchStage := bson.M{}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchStage["branchId"] = branchObjId
	}

	return []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id": bson.M{
				"productId": "$productId",
				"unitId":    "$unitId",
			},
			"totalStock": bson.M{"$sum": "$quantity"},
			"totalCost":  bson.M{"$sum": bson.M{"$multiply": bson.A{"$quantity", "$costPrice"}}},
		}},
		{"$lookup": bson.M{
			"from":         "products",
			"localField":   "_id.productId",
			"foreignField": "_id",
			"as":           "product",
		}},
		{"$unwind": "$product"},
		{"$lookup": bson.M{
			"from":         "product_units",
			"localField":   "_id.unitId",
			"foreignField": "_id",
			"as":           "unit",
		}},
		{"$unwind": "$unit"},
		{"$project": bson.M{
			"_id":          "$_id.productId",
			"totalStock":   1,
			"totalCost":    1,
			"name":         "$product.name",
			"serialNumber": "$product.serialNumber",
			"unit":         "$unit.unit",
		}},
		{"$sort": bson.M{"name": 1, "unit": 1}},
	}, nil
}

func (entity *productStockEntity) GetDeadStockProducts(days int, branchId string) ([]entities.DeadStockProduct, error) {
	logrus.Info("GetDeadStockProducts")
	ctx, cancel := utils.InitContext()
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -days)
	pipeline, err := buildDeadStockProductsPipeline(cutoff, branchId)
	if err != nil {
		return nil, err
	}

	var results []entities.DeadStockProduct
	cursor, err := entity.productStockRepo.Aggregate(ctx, pipeline)
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

func buildDeadStockProductsPipeline(cutoff time.Time, branchId string) ([]bson.M, error) {
	matchFilter := bson.M{"quantity": bson.M{"$gt": 0}}
	orderLookupMatch := bson.A{
		bson.M{"$eq": bson.A{"$productId", "$$pid"}},
		bson.M{"$eq": bson.A{"$unitId", "$$uid"}},
	}
	if branchId != "" {
		branchObjId, err := primitive.ObjectIDFromHex(branchId)
		if err != nil {
			return nil, err
		}
		matchFilter["branchId"] = branchObjId
		orderLookupMatch = append(orderLookupMatch, bson.M{"$eq": bson.A{"$branchId", branchObjId}})
	}

	return []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id": bson.M{
				"productId": "$productId",
				"unitId":    "$unitId",
			},
			"quantity":  bson.M{"$sum": "$quantity"},
			"totalCost": bson.M{"$sum": bson.M{"$multiply": bson.A{"$quantity", "$costPrice"}}},
		}},
		{"$lookup": bson.M{
			"from": "products",
			"let":  bson.M{"pid": "$_id.productId"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$pid"}}, "deletedDate": bson.M{"$exists": false}}},
				bson.M{"$project": bson.M{"name": 1}},
			},
			"as": "product",
		}},
		{"$unwind": "$product"},
		{"$lookup": bson.M{
			"from":         "product_units",
			"localField":   "_id.unitId",
			"foreignField": "_id",
			"as":           "unit",
		}},
		{"$unwind": "$unit"},
		{"$lookup": bson.M{
			"from": "order_items",
			"let":  bson.M{"pid": "$_id.productId", "uid": "$_id.unitId"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": orderLookupMatch}}},
				bson.M{"$lookup": bson.M{
					"from": "orders",
					"let":  bson.M{"oid": "$orderId"},
					"pipeline": bson.A{
						bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$_id", "$$oid"}},
							bson.M{"$in": bson.A{"$status", bson.A{constant.ACTIVE, constant.CONFIRMED}}},
						}}}},
						bson.M{"$project": bson.M{"_id": 1}},
					},
					"as": "order",
				}},
				bson.M{"$unwind": "$order"},
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
			"costPrice": bson.M{"$cond": bson.M{
				"if":   bson.M{"$gt": bson.A{"$quantity", 0}},
				"then": bson.M{"$divide": bson.A{"$totalCost", "$quantity"}},
				"else": 0,
			}},
		}},
		{"$match": bson.M{
			"$or": bson.A{
				bson.M{"lastSold": nil},
				bson.M{"lastSold": bson.M{"$lt": cutoff}},
			},
		}},
		{"$project": bson.M{
			"_id":       bson.M{"$toString": "$_id.productId"},
			"name":      "$product.name",
			"unit":      "$unit.unit",
			"quantity":  1,
			"costPrice": 1,
			"lastSold": bson.M{"$cond": bson.M{
				"if":   bson.M{"$eq": bson.A{"$lastSold", nil}},
				"then": "",
				"else": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$lastSold"}},
			}},
		}},
		{"$sort": bson.M{"lastSold": 1, "name": 1, "unit": 1}},
	}, nil
}
