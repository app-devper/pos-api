package legacy

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// missingId shows an offender whose _id is missing or null: a Line grouped by
// an orderId it does not have.
const missingId = "(missing)"

// offenders is one finding: the documents of a pipeline that match match.
type offenders struct {
	kind, detail string
	match        bson.M
}

// idsFacet keeps the first maxFindingIds matching _id values, smallest first.
func idsFacet(match bson.M) bson.A {
	return bson.A{
		bson.M{"$match": match},
		bson.M{"$sort": bson.M{"_id": 1}},
		bson.M{"$limit": maxFindingIds},
		bson.M{"$project": bson.M{"_id": 1}},
	}
}

// totalFacet counts the matching documents.
func totalFacet(match bson.M) bson.A {
	return bson.A{bson.M{"$match": match}, bson.M{"$count": "n"}}
}

func idsKey(i int) string   { return fmt.Sprintf("ids%d", i) }
func totalKey(i int) string { return fmt.Sprintf("total%d", i) }

// facetOf splits the documents of a pipeline into groups, each capped and
// counted by the database: the result is one small document.
func facetOf(groups []offenders) bson.D {
	facets := bson.D{}
	for i, g := range groups {
		facets = append(facets, bson.E{Key: idsKey(i), Value: idsFacet(g.match)}, bson.E{Key: totalKey(i), Value: totalFacet(g.match)})
	}
	return bson.D{{Key: "$facet", Value: facets}}
}

// findOffenders runs pipeline on coll and returns one finding per group with
// offenders, in the order of groups. Only maxFindingIds ids per group leave the
// database.
func findOffenders(ctx context.Context, db *mongo.Database, coll string, pipeline mongo.Pipeline, groups []offenders) ([]Finding, error) {
	full := append(append(mongo.Pipeline{}, pipeline...), facetOf(groups))
	cursor, err := db.Collection(coll).Aggregate(ctx, full, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, fmt.Errorf("check %s on %s: %w", groups[0].kind, coll, err)
	}
	defer cursor.Close(ctx)
	var rows []bson.Raw
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("read %s on %s: %w", groups[0].kind, coll, err)
	}
	if len(rows) != 1 {
		return nil, fmt.Errorf("read %s on %s: %d facet results, want 1", groups[0].kind, coll, len(rows))
	}
	found := []Finding{}
	for i, g := range groups {
		f, err := findingAt(rows[0], i, g)
		if err != nil {
			return nil, fmt.Errorf("read %s on %s: %w", g.kind, coll, err)
		}
		found = append(found, f...)
	}
	return found, nil
}

// findingAt reads group i of a facet result; none when it has no offender.
func findingAt(row bson.Raw, i int, g offenders) ([]Finding, error) {
	var ids []struct {
		Id bson.RawValue `bson:"_id"`
	}
	if err := row.Lookup(idsKey(i)).Unmarshal(&ids); err != nil {
		return nil, err
	}
	var total []struct {
		N int64 `bson:"n"`
	}
	if err := row.Lookup(totalKey(i)).Unmarshal(&total); err != nil {
		return nil, err
	}
	if len(total) == 0 || total[0].N == 0 {
		return nil, nil
	}
	hex := make([]string, 0, len(ids))
	for _, id := range ids {
		hex = append(hex, idText(id.Id))
	}
	return []Finding{{Kind: g.kind, Detail: g.detail, Ids: hex, Total: total[0].N}}, nil
}

// idText shows an _id: an ObjectID as hex, a missing or null one as
// missingId, anything else (a text orderId, say) in its extended JSON form.
func idText(id bson.RawValue) string {
	if oid, ok := id.ObjectIDOK(); ok {
		return oid.Hex()
	}
	if id.Type == 0 || id.Type == bson.TypeNull || id.Type == bson.TypeUndefined {
		return missingId
	}
	return id.String()
}
