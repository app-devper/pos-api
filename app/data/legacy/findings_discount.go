package legacy

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// discountKind is one way an Order's stored total can disagree with the
// per-unit discount the current code applies, in the order findings reports
// them.
type discountKind struct {
	kind, detail, sum string
}

var discountKinds = []discountKind{
	{KindDiscountPerLine, "Orders whose total matches a per-line discount, not the per-unit discount the current code applies", "$perLine"},
	{KindDiscountNotSubtracted, "Orders whose total is the plain sum of line prices, discount not subtracted (the 2024 server's own recompute)", "$plain"},
	{KindDiscountUnknown, "Orders with a discounted line whose total matches none of the three discount rules", ""},
}

// field returns the line value or 0 when it is missing.
func field(name string) bson.M {
	return bson.M{"$ifNull": bson.A{"$" + name, 0}}
}

// orderSums groups every line by its Order once: whether any line carries a
// discount, and Σ price under each rule. A line's price is its whole amount
// before discount. Grouping all lines (not only the discounted ones) keeps the
// undiscounted lines of a discounted Order in its sums.
func orderSums() bson.D {
	discount := field("discount")
	return bson.D{{Key: "$group", Value: bson.M{
		"_id":        "$orderId",
		"discounted": bson.M{"$max": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{discount, 0}}, 1, 0}}},
		"perUnit":    bson.M{"$sum": bson.M{"$subtract": bson.A{field("price"), bson.M{"$multiply": bson.A{discount, field("quantity")}}}}},
		"perLine":    bson.M{"$sum": bson.M{"$subtract": bson.A{field("price"), discount}}},
		"plain":      bson.M{"$sum": field("price")},
	}}}
}

// near is true when the Order's total is within amountTolerance of sum.
func near(sum string) bson.M {
	diff := bson.M{"$abs": bson.M{"$subtract": bson.A{"$total", sum}}}
	return bson.M{"$lte": bson.A{diff, amountTolerance}}
}

// kindOf names the first discount kind an Order matches; "" when its total
// follows the current per-unit rule, whatever else it also matches.
func kindOf() bson.M {
	branches := bson.A{bson.M{"case": near("$perUnit"), "then": ""}}
	for _, k := range discountKinds {
		if k.sum != "" {
			branches = append(branches, bson.M{"case": near(k.sum), "then": k.kind})
		}
	}
	return bson.M{"$switch": bson.M{"branches": branches, "default": KindDiscountUnknown}}
}

// discountPipeline starts from order_items, so it needs no index on orderId:
// one collection scan, then each Order is found by its _id index.
func discountPipeline() mongo.Pipeline {
	return mongo.Pipeline{
		orderSums(),
		{{Key: "$match", Value: bson.M{"discounted": 1}}},
		{{Key: "$lookup", Value: bson.M{"from": "orders", "localField": "_id", "foreignField": "_id", "as": "order"}}},
		{{Key: "$unwind", Value: "$order"}},
		{{Key: "$addFields", Value: bson.M{"total": bson.M{"$ifNull": bson.A{"$order.total", 0}}}}},
		{{Key: "$project", Value: bson.M{"kind": kindOf()}}},
		{{Key: "$match", Value: bson.M{"kind": bson.M{"$ne": ""}}}},
	}
}

// checkDiscountConvention reports, in one pass, the Orders with a discounted
// line whose total does not follow the current per-unit discount.
func checkDiscountConvention(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	cursor, err := db.Collection("order_items").Aggregate(ctx, discountPipeline(), options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, fmt.Errorf("check discounts on order_items: %w", err)
	}
	defer cursor.Close(ctx)
	var rows []struct {
		Id   primitive.ObjectID `bson:"_id"`
		Kind string             `bson:"kind"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("read discounts on order_items: %w", err)
	}
	byKind := map[string][]string{}
	for _, row := range rows {
		byKind[row.Kind] = append(byKind[row.Kind], row.Id.Hex())
	}
	found := []Finding{}
	for _, k := range discountKinds {
		found = append(found, findingOf(k.kind, k.detail, byKind[k.kind])...)
	}
	return found, nil
}
