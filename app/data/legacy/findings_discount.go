package legacy

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// numericTypes are the BSON types $isNumber accepts, and the types that leave
// a line amount at 0 without being reported (missing, null).
var numericTypes = bson.A{"double", "int", "long", "decimal", "missing", "null"}

// lineAmounts are the line fields the sums read.
var lineAmounts = []string{"price", "quantity", "discount"}

// field returns the value at path when it is a number, otherwise 0: a missing,
// null or non-numeric amount neither stops the aggregation nor compares as
// greater than 0.
func field(path string) bson.M {
	return bson.M{"$cond": bson.A{bson.M{"$isNumber": "$" + path}, "$" + path, 0}}
}

// nonNumericLine is 1 when a line amount holds a value that is not a number
// (a missing or null amount is not counted), otherwise 0.
func nonNumericLine() bson.M {
	anyAmount := bson.A{}
	for _, name := range lineAmounts {
		anyAmount = append(anyAmount, bson.M{"$not": bson.A{bson.M{"$in": bson.A{bson.M{"$type": "$" + name}, numericTypes}}}})
	}
	return bson.M{"$cond": bson.A{bson.M{"$or": anyAmount}, 1, 0}}
}

// orderSums groups every line by its Order once: whether any line carries a
// discount or a non-numeric amount, and Σ price under each rule. A line's
// price is its whole amount before discount. Grouping all lines (not only the
// discounted ones) keeps the undiscounted lines of a discounted Order in its
// sums.
func orderSums() bson.D {
	discount := field("discount")
	return bson.D{{Key: "$group", Value: bson.M{
		"_id":        "$orderId",
		"discounted": bson.M{"$max": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{discount, 0}}, 1, 0}}},
		"nonNumeric": bson.M{"$max": nonNumericLine()},
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
// one collection scan, then each Order is found by its _id index. A row is an
// Order with a discount kind, a non-numeric line amount, or both.
func discountPipeline() mongo.Pipeline {
	hasDiscountAndOrder := bson.M{"$and": bson.A{
		bson.M{"$eq": bson.A{"$discounted", 1}},
		bson.M{"$ne": bson.A{bson.M{"$type": "$order"}, "missing"}},
	}}
	return mongo.Pipeline{
		orderSums(),
		{{Key: "$match", Value: bson.M{"$or": bson.A{bson.M{"discounted": 1}, bson.M{"nonNumeric": 1}}}}},
		{{Key: "$lookup", Value: bson.M{"from": "orders", "localField": "_id", "foreignField": "_id", "as": "order"}}},
		{{Key: "$unwind", Value: bson.M{"path": "$order", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$addFields", Value: bson.M{"total": field("order.total")}}},
		{{Key: "$project", Value: bson.M{"nonNumeric": 1, "kind": bson.M{"$cond": bson.A{hasDiscountAndOrder, kindOf(), ""}}}}},
		{{Key: "$match", Value: bson.M{"$or": bson.A{bson.M{"kind": bson.M{"$ne": ""}}, bson.M{"nonNumeric": 1}}}}},
	}
}

// nonNumericDetail explains KindNonNumericLineAmount.
const nonNumericDetail = "Orders with a Line whose price, quantity or discount is not a number; the discount check counted it as 0"

// checkDiscountConvention reports, in one pass, the Orders with a discounted
// line whose total does not follow the current per-unit discount, then the
// Orders with a non-numeric line amount.
func checkDiscountConvention(ctx context.Context, db *mongo.Database) ([]Finding, error) {
	groups := make([]offenders, 0, len(discountKinds)+1)
	for _, k := range discountKinds {
		groups = append(groups, offenders{kind: k.kind, detail: k.detail, match: bson.M{"kind": k.kind}})
	}
	groups = append(groups, offenders{kind: KindNonNumericLineAmount, detail: nonNumericDetail, match: bson.M{"nonNumeric": 1}})
	return findOffenders(ctx, db, "order_items", discountPipeline(), groups)
}
