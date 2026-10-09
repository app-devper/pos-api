// Package catalogue holds the rules of a Product's catalogue that more than
// one module reads: the repositories that edit Units, and the Stock ledger
// that receives into them.
package catalogue

import (
	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson"
)

// MainUnit matches a Product's main Unit: the Unit named as the Product's
// unit. A Receive is entered in it, so it keeps its name and size and is
// never removed (see repositories.ErrMainUnitFixed).
func MainUnit(product entities.Product) bson.M {
	return bson.M{"productId": product.Id, "unit": product.Unit}
}
