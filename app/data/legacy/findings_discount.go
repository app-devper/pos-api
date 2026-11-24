package legacy

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// lineSum is Σ over the lines of price − per×(multiplier), as an aggregation
// expression; per is the line's discount multiplied by quantity or by 1.
func lineSum(perUnit bool) bson.M {
	discount := bson.M{"$ifNull": bson.A{"$$l.discount", 0}}
	if perUnit {
		discount = bson.M{"$multiply": bson.A{discount, bson.M{"$ifNull": bson.A{"$$l.quantity", 0}}}}
	}
	return bson.M{"$sum": bson.M{"$map": bson.M{
		"input": "$lines", "as": "l",
		"in": bson.M{"$subtract": bson.A{bson.M{"$ifNull": bson.A{"$$l.price", 0}}, discount}},
	}}}
}

// discountedOrders yields the Orders with a discounted line, with their total
// compared to both conventions: perUnit (current code) and perLine (2024 till).
func discountedOrders() mongo.Pipeline {
	diff := func(sum bson.M) bson.M {
		return bson.M{"$abs": bson.M{"$subtract": bson.A{bson.M{"$ifNull": bson.A{"$total", 0}}, sum}}}
	}
	return mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{"from": "order_items", "localField": "_id", "foreignField": "orderId", "as": "lines"}}},
		{{Key: "$match", Value: bson.M{"lines.discount": bson.M{"$gt": 0}}}},
		{{Key: "$addFields", Value: bson.M{
			"perUnitDiff": diff(lineSum(true)),
			"perLineDiff": diff(lineSum(false)),
		}}},
	}
}

func withStages(base mongo.Pipeline, stages ...bson.D) mongo.Pipeline {
	out := make(mongo.Pipeline, 0, len(base)+len(stages))
	out = append(out, base...)
	return append(out, stages...)
}

// checkDiscountConvention reports Orders whose total does not follow the
// current per-unit discount: matching only the per-line one is the 2024 till;
// matching neither is unknown. An Order matching both is fine.
func checkDiscountConvention(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	off := bson.M{"$gt": amountTolerance}
	on := bson.M{"$lte": amountTolerance}
	project := bson.D{{Key: "$project", Value: bson.M{"_id": 1}}}
	perLine := withStages(discountedOrders(),
		bson.D{{Key: "$match", Value: bson.M{"perUnitDiff": off, "perLineDiff": on}}}, project)
	unknown := withStages(discountedOrders(),
		bson.D{{Key: "$match", Value: bson.M{"perUnitDiff": off, "perLineDiff": off}}}, project)

	first, err := idFinding(ctx, db, "orders", perLine, KindDiscountPerLine,
		"Orders whose total matches a per-line discount, not the per-unit discount the current code applies")
	if err != nil {
		return nil, err
	}
	second, err := idFinding(ctx, db, "orders", unknown, KindDiscountUnknown,
		"Orders with a discounted line whose total matches neither discount convention")
	if err != nil {
		return nil, err
	}
	return append(first, second...), nil
}
