package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A Product's main Unit is the one named as the Product's unit: Receives are
// entered in it, so it keeps its name and size, and is never removed.
type catalogueFixture struct {
	t        *testing.T
	pos      *mongo.Database
	products *productEntity
	product  primitive.ObjectID
	main     primitive.ObjectID
	box      primitive.ObjectID
}

func newCatalogueFixture(t *testing.T) *catalogueFixture {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	pos := client.Database(fmt.Sprintf("pos_catalogue_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	f := &catalogueFixture{t: t, pos: pos, products: NewProductEntity(&db.Resource{Client: client, PosDb: pos}).(*productEntity),
		product: primitive.NewObjectID(), main: primitive.NewObjectID(), box: primitive.NewObjectID()}
	f.insert("products", bson.M{"_id": f.product, "name": "Paracetamol", "unit": "TAB"})
	f.insert("product_units", entities.ProductUnit{Id: f.main, ProductId: f.product, Unit: "TAB", Size: 1})
	f.insert("product_units", entities.ProductUnit{Id: f.box, ProductId: f.product, Unit: "BOX", Size: 10})
	f.insert("product_prices", entities.ProductPrice{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.box, CustomerType: "General", Price: 100})
	return f
}

func (f *catalogueFixture) insert(collection string, doc any) {
	f.t.Helper()
	if _, err := f.pos.Collection(collection).InsertOne(context.Background(), doc); err != nil {
		f.t.Fatal(err)
	}
}

func (f *catalogueFixture) count(collection string, filter bson.M) int64 {
	f.t.Helper()
	n, err := f.pos.Collection(collection).CountDocuments(context.Background(), filter)
	if err != nil {
		f.t.Fatal(err)
	}
	return n
}

func unitEdit(unit string, size int) request.ProductUnit {
	return request.ProductUnit{Unit: unit, Size: size, CostPrice: 1}
}

func TestTheMainUnitKeepsItsNameAndSize(t *testing.T) {
	for name, edit := range map[string]request.ProductUnit{
		"rename": unitEdit("PILL", 1),
		"resize": unitEdit("TAB", 2),
	} {
		t.Run(name, func(t *testing.T) {
			f := newCatalogueFixture(t)
			_, err := f.products.UpdateProductUnitById(f.main.Hex(), edit)
			if !errors.Is(err, ErrMainUnitFixed) {
				t.Fatalf("got %v, want ErrMainUnitFixed", err)
			}
			if f.count("product_units", bson.M{"_id": f.main, "unit": "TAB", "size": 1}) != 1 {
				t.Fatal("the main Unit changed")
			}
		})
	}
}

func TestTheMainUnitCanChangeWhatDoesNotDefineIt(t *testing.T) {
	f := newCatalogueFixture(t)
	edit := unitEdit("TAB", 1)
	edit.Barcode = "885000"
	if _, err := f.products.UpdateProductUnitById(f.main.Hex(), edit); err != nil {
		t.Fatal(err)
	}
	if f.count("product_units", bson.M{"_id": f.main, "barcode": "885000"}) != 1 {
		t.Fatal("barcode not saved")
	}
}

func TestAUnitWithStockKeepsItsSize(t *testing.T) {
	f := newCatalogueFixture(t)
	f.insert("product_stocks", entities.ProductStock{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.box, BranchId: primitive.NewObjectID(), Quantity: 0})

	if _, err := f.products.UpdateProductUnitById(f.box.Hex(), unitEdit("BOX", 12)); !errors.Is(err, ErrUnitSizeFixed) {
		t.Fatalf("got %v, want ErrUnitSizeFixed", err)
	}
	if _, err := f.products.UpdateProductUnitById(f.box.Hex(), unitEdit("CARTON", 10)); err != nil {
		t.Fatalf("renaming a Unit that is not the main Unit: %v", err)
	}
}

func TestWhichUnitsCanBeRemoved(t *testing.T) {
	cases := []struct {
		name  string
		setup func(f *catalogueFixture)
		unit  func(f *catalogueFixture) primitive.ObjectID
		want  error
	}{
		{"the main Unit, even after its size moved off 1", func(f *catalogueFixture) {
			_, _ = f.pos.Collection("product_units").UpdateOne(context.Background(), bson.M{"_id": f.main}, bson.M{"$set": bson.M{"size": 3}})
		}, func(f *catalogueFixture) primitive.ObjectID { return f.main }, ErrMainUnitFixed},
		{"a Unit with Stock", func(f *catalogueFixture) {
			f.insert("product_stocks", entities.ProductStock{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.box, BranchId: primitive.NewObjectID()})
		}, func(f *catalogueFixture) primitive.ObjectID { return f.box }, ErrUnitInUse},
		{"a Unit a Line still owes an Oversell of", func(f *catalogueFixture) {
			f.insert("order_items", entities.OrderItem{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.box, Quantity: 2, OversoldQty: 2, Status: constant.CONFIRMED})
		}, func(f *catalogueFixture) primitive.ObjectID { return f.box }, ErrUnitInUse},
		{"a Unit nothing uses", func(f *catalogueFixture) {
			f.insert("order_items", entities.OrderItem{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.box, Quantity: 2, OversoldQty: 2, Status: constant.CANCELLED})
		}, func(f *catalogueFixture) primitive.ObjectID { return f.box }, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newCatalogueFixture(t)
			c.setup(f)
			unit := c.unit(f)

			_, err := f.products.RemoveProductUnitCascade(unit.Hex(), "", "u1")

			if c.want == nil {
				if err != nil {
					t.Fatal(err)
				}
				if f.count("product_units", bson.M{"_id": unit}) != 0 || f.count("product_prices", bson.M{"unitId": unit}) != 0 {
					t.Fatal("the Unit or its prices remain")
				}
				return
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
			if f.count("product_units", bson.M{"_id": unit}) != 1 {
				t.Fatal("the Unit was removed")
			}
		})
	}
}
