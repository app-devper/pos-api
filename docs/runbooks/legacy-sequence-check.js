// Prints every sequences row and, per counter, the highest code already used,
// so the surviving row of a DUPLICATE_SEQUENCE can be checked
// (docs/runbooks/legacy-db-findings.md, "Duplicate sequence rows"). Codes are
// compared by date part, then by number, not as text: OD_2026112410000 sorts
// above OD_202611249999. Numbers are read as decimals, so a code of any
// length (up to 34 digits) compares right. Run on the COPY only. Reads only.
// Usage: COPY_DB=<COPY_DB_NAME> mongosh '<NEW_MONGO_URI>' --quiet legacy-sequence-check.js
const name = process.env.COPY_DB;
if (!name) { print('set COPY_DB to the copy database name'); quit(2); }
const d = db.getSiblingDB(name);
d.sequences.find({}, {field: 1, type: 1, date: 1, value: 1}).sort({field: 1, _id: 1})
  .forEach(r => print('row', r._id.toHexString(), r.field, r.type, 'date', r.date, 'value', String(r.value)));
// field: counter; coll/key: where its codes are kept; prefix; dateLen: digits of date in the code
const series = [
  {field: 'ORDER', coll: 'orders', key: 'code', prefix: 'OD_', dateLen: 8},
  {field: 'RECEIVE', coll: 'receives', key: 'code', prefix: 'RC_', dateLen: 8},
  {field: 'MEMBER', coll: 'customers', key: 'code', prefix: 'MB_', dateLen: 4},
  {field: 'PRODUCT', coll: 'products', key: 'serialNumber', prefix: 'PD_', dateLen: 0},
];
for (const s of series) {
  const top = d.getCollection(s.coll).aggregate([
    {$match: {[s.key]: {$regex: '^' + s.prefix + '[0-9]{' + (s.dateLen + 1) + ',}$'}}},
    {$project: {_id: 0, code: '$' + s.key, digits: {$substrCP: ['$' + s.key, s.prefix.length, 100]}}},
    {$project: {code: 1, date: {$substrCP: ['$digits', 0, s.dateLen]},
                number: {$convert: {input: {$substrCP: ['$digits', s.dateLen, 100]}, to: 'decimal', onError: null}}}},
    {$sort: {date: -1, number: -1}},
    {$limit: 1},
  ], {allowDiskUse: true}).toArray();
  print(s.field, 'highest code', top.length ? top[0].code + ' date ' + (top[0].date || '-') + ' number ' + String(top[0].number) : 'none');
}
