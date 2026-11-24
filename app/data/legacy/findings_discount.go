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

// floatTypes decode into the float64 price and discount of entities.OrderItem
// (missing and null decode as 0); a Decimal128 or any non-number does not.
var floatTypes = bson.A{"double", "int", "long", "missing", "null"}

// intTypes decode into its int quantity; a double does only when it is whole
// and in range, which wholeDouble checks.
var intTypes = bson.A{"int", "long", "missing", "null"}

// maxWholeDouble bounds a double the driver puts into an int64 without
// overflow; NaN compares below every number in MongoDB, so it fails the
// lower bound.
const maxWholeDouble = 9.2e18

// field returns the value at path when it is a number, otherwise 0: a missing,
// null or non-numeric amount neither stops the aggregation nor compares as
// greater than 0.
func field(path string) bson.M {
	return bson.M{"$cond": bson.A{bson.M{"$isNumber": "$" + path}, "$" + path, 0}}
}

// typeIn is true when the value at path has one of types.
func typeIn(path string, types bson.A) bson.M {
	return bson.M{"$in": bson.A{bson.M{"$type": "$" + path}, types}}
}

// wholeDouble is true for a double quantity with no fraction, in range. $and
// stops at the first false, so $trunc only sees a double.
func wholeDouble(path string) bson.M {
	v := "$" + path
	return bson.M{"$and": bson.A{
		bson.M{"$eq": bson.A{bson.M{"$type": v}, "double"}},
		bson.M{"$gte": bson.A{v, -maxWholeDouble}},
		bson.M{"$lte": bson.A{v, maxWholeDouble}},
		bson.M{"$eq": bson.A{v, bson.M{"$trunc": v}}},
	}}
}

// undecodableLine is 1 when the current code could not read a line amount
// into entities.OrderItem (a text, a Decimal128, a fractional or out-of-range
// quantity), otherwise 0.
func undecodableLine() bson.M {
	readable := bson.M{"$and": bson.A{
		typeIn("price", floatTypes),
		typeIn("discount", floatTypes),
		bson.M{"$or": bson.A{typeIn("quantity", intTypes), wholeDouble("quantity")}},
	}}
	return bson.M{"$cond": bson.A{readable, 0, 1}}
}

// orderSums groups every line by its Order once: whether any line carries a
// discount or an amount the current code cannot read, and Σ price under each rule. A line's
// price is its whole amount before discount. Grouping all lines (not only the
// discounted ones) keeps the undiscounted lines of a discounted Order in its
// sums.
func orderSums() bson.D {
	discount := field("discount")
	return bson.D{{Key: "$group", Value: bson.M{
		"_id":        "$orderId",
		"discounted": bson.M{"$max": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{discount, 0}}, 1, 0}}},
		"nonNumeric": bson.M{"$max": undecodableLine()},
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
const nonNumericDetail = "Orders with a Line whose price, quantity or discount the current code cannot read: not a number (counted as 0 by the discount check), a Decimal128, or a fractional quantity"

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
