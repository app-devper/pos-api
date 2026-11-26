package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
)

// namedCountLines resolves, per Count, the names of its Lines' Products, Units
// and Stocks: one lookup per collection for every Count the read returns, so
// counts recorded before names were returned are named too. A Product keeps
// its document when soft-deleted, so it keeps its name; a Line whose Product
// or Stock has gone is returned with empty names.
func namedCountLines() []bson.M {
	return []bson.M{
		lookupByIds("products", "$items.productId", "_products", bson.M{"name": 1}),
		lookupByIds("product_stocks", "$items.stockId", "_stocks", bson.M{"lotNumber": 1, "unitId": 1}),
		lookupByIds("product_units", "$_stocks.unitId", "_units", bson.M{"unit": 1}),
		{"$addFields": bson.M{"items": bson.M{"$map": bson.M{
			"input": bson.M{"$ifNull": bson.A{"$items", bson.A{}}},
			"as":    "line",
			"in":    namedLine(),
		}}}},
		{"$project": bson.M{"_products": 0, "_stocks": 0, "_units": 0}},
	}
}

// lookupByIds joins the documents of a collection whose _id is in ids,
// keeping only the fields given.
func lookupByIds(from string, ids string, as string, fields bson.M) bson.M {
	return bson.M{"$lookup": bson.M{
		"from": from,
		"let":  bson.M{"ids": bson.M{"$ifNull": bson.A{ids, bson.A{}}}},
		"pipeline": bson.A{
			bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$ids"}}}},
			bson.M{"$project": fields},
		},
		"as": as,
	}}
}

// firstWithId is the first element of the joined array whose _id is id
// (missing when there is none). It sticks to operators older servers know.
func firstWithId(array string, id any) bson.M {
	return bson.M{"$arrayElemAt": bson.A{bson.M{"$filter": bson.M{
		"input": array, "as": "doc", "cond": bson.M{"$eq": bson.A{"$$doc._id", id}},
	}}, 0}}
}

// namedLine is one stored Line ($$line) with productName, unit and lotNumber.
func namedLine() bson.M {
	return bson.M{"$let": bson.M{
		"vars": bson.M{
			"product": firstWithId("$_products", "$$line.productId"),
			"stock":   firstWithId("$_stocks", "$$line.stockId"),
		},
		"in": bson.M{"$let": bson.M{
			"vars": bson.M{"unit": firstWithId("$_units", "$$stock.unitId")},
			"in": bson.M{"$mergeObjects": bson.A{"$$line", bson.M{
				"productName": bson.M{"$ifNull": bson.A{"$$product.name", ""}},
				"unit":        bson.M{"$ifNull": bson.A{"$$unit.unit", ""}},
				"lotNumber":   bson.M{"$ifNull": bson.A{"$$stock.lotNumber", ""}},
			}}},
		}},
	}}
}
