package repositories

import (
	"context"

	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// countNamer are the collections a Count's Lines are named from.
type countNamer struct {
	products, stocks, units *mongo.Collection
}

type lineProduct struct {
	Id   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}

type lineStock struct {
	Id        primitive.ObjectID `bson:"_id"`
	UnitId    primitive.ObjectID `bson:"unitId"`
	LotNumber string             `bson:"lotNumber"`
}

type lineUnit struct {
	Id   primitive.ObjectID `bson:"_id"`
	Unit string             `bson:"unit"`
}

// name returns counts with every Line named at read time: one find by _id per
// collection for all the counts together, so counts recorded before names
// were returned are named too. A Product keeps its document when
// soft-deleted, so it keeps its name; a Line whose Product or Stock has gone
// gets empty names.
func (n countNamer) name(ctx context.Context, counts []entities.StockCount) ([]entities.StockCountDetail, error) {
	productIds, stockIds := lineIds(counts)
	var products []lineProduct
	if err := findByIds(ctx, n.products, productIds, bson.M{"name": 1}, &products); err != nil {
		return nil, err
	}
	var stocks []lineStock
	if err := findByIds(ctx, n.stocks, stockIds, bson.M{"unitId": 1, "lotNumber": 1}, &stocks); err != nil {
		return nil, err
	}
	unitIds := make([]primitive.ObjectID, 0, len(stocks))
	for _, st := range stocks {
		unitIds = append(unitIds, st.UnitId)
	}
	var units []lineUnit
	if err := findByIds(ctx, n.units, distinct(unitIds), bson.M{"unit": 1}, &units); err != nil {
		return nil, err
	}
	return namedCounts(counts, products, stocks, units), nil
}

// lineIds is the distinct Product and Stock ids over every Line of counts.
func lineIds(counts []entities.StockCount) ([]primitive.ObjectID, []primitive.ObjectID) {
	var products, stocks []primitive.ObjectID
	for _, c := range counts {
		for _, line := range c.Items {
			products = append(products, line.ProductId)
			stocks = append(stocks, line.StockId)
		}
	}
	return distinct(products), distinct(stocks)
}

func distinct(ids []primitive.ObjectID) []primitive.ObjectID {
	seen := map[primitive.ObjectID]bool{}
	out := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// findByIds decodes into out the documents whose _id is in ids, with only
// the fields given; no ids, no query.
func findByIds(ctx context.Context, col *mongo.Collection, ids []primitive.ObjectID, fields bson.M, out any) error {
	if len(ids) == 0 {
		return nil
	}
	cursor, err := col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, options.Find().SetProjection(fields))
	if err != nil {
		return err
	}
	return cursor.All(ctx, out)
}

// namedCounts builds the read shape of counts from what was found; the
// loaded counts are left as they are.
func namedCounts(counts []entities.StockCount, products []lineProduct, stocks []lineStock, units []lineUnit) []entities.StockCountDetail {
	productName := map[primitive.ObjectID]string{}
	for _, p := range products {
		productName[p.Id] = p.Name
	}
	unitName := map[primitive.ObjectID]string{}
	for _, u := range units {
		unitName[u.Id] = u.Unit
	}
	stockById := map[primitive.ObjectID]lineStock{}
	for _, st := range stocks {
		stockById[st.Id] = st
	}
	details := make([]entities.StockCountDetail, len(counts))
	for i, c := range counts {
		lines := make([]entities.StockCountLine, len(c.Items))
		for j, item := range c.Items {
			st, found := stockById[item.StockId]
			line := entities.StockCountLine{StockCountItem: item, ProductName: productName[item.ProductId], LotNumber: st.LotNumber}
			if found {
				line.Unit = unitName[st.UnitId]
			}
			lines[j] = line
		}
		details[i] = entities.StockCountDetail{Id: c.Id, BranchId: c.BranchId, CountNo: c.CountNo, Note: c.Note,
			Items: lines, CreatedBy: c.CreatedBy, CreatedDate: c.CreatedDate}
	}
	return details
}
