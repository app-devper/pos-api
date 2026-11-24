// Prints the counts that change when the legacy till or its back office is
// used. Run it on the legacy database and on the copy, then diff the outputs
// (docs/runbooks/legacy-db-migration.md, step 6.3). Reads only.
// Usage: COMPARE_DB=<database name> mongosh '<uri>' --quiet legacy-db-compare.js
const name = process.env.COMPARE_DB;
if (!name) { print('set COMPARE_DB to the database name'); quit(2); }
const d = db.getSiblingDB(name);
for (const c of ['orders', 'order_items', 'payments', 'receives', 'product_histories', 'product_stocks',
                 'products', 'product_units', 'product_prices', 'customers']) {
  print(c, d.getCollection(c).countDocuments({}));
}
const qty = d.product_stocks.aggregate([{$group: {_id: null, q: {$sum: '$quantity'}}}]).toArray();
print('product_stocks quantity sum', qty.length ? qty[0].q : 0);
print('newest order createdDate', JSON.stringify(d.orders.find({}, {createdDate: 1, _id: 0}).sort({createdDate: -1}).limit(1).toArray()));
print('sequences', JSON.stringify(d.sequences.find({}, {_id: 0, field: 1, date: 1, value: 1}).sort({field: 1}).toArray()));
