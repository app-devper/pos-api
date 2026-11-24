// Lists every Order that has a Line with discount > 0, run on the COPY only
// (docs/runbooks/legacy-db-migration.md, "Discounted 2024 Orders"). Reads only.
// Usage: COPY_DB=<COPY_DB_NAME> mongosh '<NEW_MONGO_URI>' --quiet legacy-discounted-orders.js
const name = process.env.COPY_DB;
if (!name) { print('set COPY_DB to the copy database name'); quit(2); }
const d = db.getSiblingDB(name);
d.order_items.aggregate([
  {$match: {discount: {$gt: 0}}},
  {$group: {_id: '$orderId'}},
  {$lookup: {from: 'orders', localField: '_id', foreignField: '_id', as: 'o'}},
  {$unwind: '$o'},
  {$project: {_id: 1, code: '$o.code', createdDate: '$o.createdDate'}},
  {$sort: {createdDate: 1}},
], {allowDiskUse: true}).forEach(o => print(o._id.toHexString(), o.code, o.createdDate && o.createdDate.toISOString()));
