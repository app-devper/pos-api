# Moving the shop from the 2024 build to the current API

For the shop owner's developer. The shop still runs a 2024 build of this API
(commit `b165b4d`, Cloud Run `pos-002-api` / `pos-003-api`, reached through
`devper-pos.web.app`) on its own MongoDB database. The current API
(`pos-dev-api`, till at `devper.web.app`) cannot read that database as it is:
every document needs a Branch. This runbook moves the data onto a copy and
switches the shop over.

> **Nothing here has been tried against real data.** The command and its checks
> are covered by tests on empty and synthetic databases only. The first run on
> the shop's data is a report run.

## Hard rules

1. **Only a restored COPY.** `migrate-legacy-db` is only ever run against a
   restored copy of the legacy database, never the database `pos-002-api` /
   `pos-003-api` use.
2. **No step is run by an agent without the owner's go-ahead.** Nothing in this
   runbook is to be executed by an agent unless the owner has said yes to that
   step, in that conversation. A yes to one step does not cover the next.
3. **The legacy services are never modified or redeployed.** Going back is
   pointing the shop at the old URL (section 7).

Placeholders used below: `<LEGACY_MONGO_URI>` (the legacy database server),
`<NEW_MONGO_URI>` (a replica set that holds the copy), `<COPY_DB_NAME>` (the
copy's name, never the legacy name). Keep real values out of this file, shell
history you share, and commits.

## 1. What must be true first

- [ ] The owner has given the go-ahead for each step below that touches the
      legacy stack (step 2.1, the dump; section 6, the comparison and the switch).
- [ ] The shop's users exist in UM under the shop's Client, with System `POS`.
      Authentication goes through UM; a user missing there cannot sign in.
- [ ] The current till (`devper.web.app`) is deployed. The 2024 till cannot sell
      against the new API: a Sale without `saleId` is refused with 400.
- [ ] `<NEW_MONGO_URI>` is a **replica set** (Atlas counts). The current API
      writes Stock in MongoDB transactions, which a standalone server refuses.
- [ ] The shop has stopped selling on the legacy till. Order codes continue from
      the `sequences` collection of the copy; a bill made on the legacy till
      after the dump is lost, and the next code can clash with it.

## 2. Dump the legacy database and restore it as a copy

Run once the shop has stopped selling (section 1, last item). This reads the
legacy database; it writes nothing to it.

1. Dump:

   ```bash
   mongodump --uri='<LEGACY_MONGO_URI>' --db=<LEGACY_DB_NAME> --out=./legacy-dump
   ```

   Expected: one `done dumping <LEGACY_DB_NAME>.<collection> (N documents)`
   line per collection, no `error` lines, exit 0.

2. Restore under a new name on the replica set:

   ```bash
   mongorestore --uri='<NEW_MONGO_URI>' \
     --nsFrom='<LEGACY_DB_NAME>.*' --nsTo='<COPY_DB_NAME>.*' ./legacy-dump
   ```

   Expected: `N document(s) restored successfully. 0 document(s) failed to
   restore.`, exit 0.

3. Compare counts. In `mongosh '<NEW_MONGO_URI>'`:

   ```js
   db.getSiblingDB('<COPY_DB_NAME>').orders.countDocuments({})
   ```

   Expected: the same number as `orders` in the legacy database. Do the same
   for `product_stocks` and `payments`.

## 3. Report run (writes nothing)

```bash
MONGO_HOST='<NEW_MONGO_URI>' MONGO_POS_DB_NAME=<COPY_DB_NAME> \
  go run ./cmd/migrate-legacy-db
```

Expected, on a copy with nothing to decide (exit 0):

```
database <COPY_DB_NAME>, mode REPORT
HQ Branch 000000000000000000000000 (created: false)
product_stocks branchId: would change <N>
orders branchId: would change <N>
order_items branchId: would change <N>
payments branchId: would change <N>
product_histories branchId: would change <N>
product_lots branchId: would change <N>
receives branchId: would change <N>
receives status: would change <N>
receives items: would change <N>
```

The all-zero HQ id means no HQ Branch exists yet; `-apply` creates it. Each
`Change` line is a field the migration sets; counts are documents found before
writing.

**Exit 3** means findings follow. Each prints as `FINDING <KIND> (<total>):
<detail>`, then up to 20 document ids. Decide each before section 4:

| Kind | Meaning | Decision |
|---|---|---|
| `NOT_REPLICA_SET` | The server is not a replica set. | Stop. Restore onto a replica set (section 2). |
| `NEGATIVE_STOCK` | Stock below zero (`product_stocks`). The current ledger refuses a move that leaves one. | Do not hand-edit `quantity`: that bypasses the Stock ledger (ADR-0001). First record, per listed row, the legacy value and the counted value. Then either (a) leave it and do a Stock count or adjustment in the new till after the move, or (b) the owner decides on a hand edit in the copy. Limitation if left: the ledger refuses a move that leaves a Stock below zero, so cancelling a bill that drew from that Stock can fail until it is corrected. |
| `DUPLICATE_SERIAL` | Products share a `serialNumber`, empty counts, soft-deleted too. The unique index on `products.serialNumber` cannot be built. | Give each a distinct serial in the copy. |
| `DUPLICATE_SEQUENCE` | `sequences` rows share a `field`. The unique index cannot be built. | See "Duplicate sequence rows" below; the right row is not the highest value. |
| `ORPHAN_STOCK` | Stock whose `productId` or `unitId` matches no Product or Unit. | Export the listed rows first (see "Before deleting orphan Stock" below). Deleting them is the owner's decision; the alternative is to restore the missing Product/Unit. |
| `NO_MAIN_UNIT` | A Product has no `product_units` row named like its own unit; a Receive is entered in the main unit. | Add the unit row in the copy. |
| `DISCOUNT_PER_LINE` | Orders whose total matches a per-line discount; the current code discounts per unit. The stored total follows the 2024 till's rule. | Decide with the owner whether the stored totals stay as they are (reports use them). The command changes nothing here. |
| `DISCOUNT_UNKNOWN` | A discounted line whose total matches neither convention. | Open the listed Orders with the owner; fix by hand in the copy, or accept. |
| `UNKNOWN_PAYMENT_TYPE` | `payments.type` outside `CASH`, `CREDIT`, `PROMPTPAY`, `TRANSFER`. | First find out what each legacy value meant (ask the shop); only then map it, or leave it. Effect of leaving it: the current Go code stores and returns `payments.type` without grouping or validating it, so the API itself is unaffected; whether the till or its reports group by it is not determined. |

### Duplicate sequence rows

`app/data/repositories/sequence.go`: each `sequences` row has `field`,
`value`, `date` (`YYYYMMDD`, shop time zone), `prefix`, `format`, `type`. The
`ORDER` (prefix `OD_`) and `RECEIVE` (prefix `RC_`) counters are `DAILY`: on
each request, if the row's `date` equals today it adds 1 to `value`, otherwise
it sets `value` to 1 and `date` to today. The code is `prefix + date +
value` zero-padded to `format` digits (4), for example `OD_202611240007`.

So the right row is the one with the **latest `date`, and among equal dates
the highest `value`**. Keeping an older date with a higher value makes today's
next order restart at 1 and collide with bills already made today. The
surviving row must cover the highest code already used in `orders` for that
date. Check in `mongosh '<NEW_MONGO_URI>'` (untried):

```js
const d = db.getSiblingDB('<COPY_DB_NAME>');
d.sequences.find({field: 'ORDER'}, {_id: 0, field: 1, date: 1, value: 1}).toArray()
d.orders.find({code: /^OD_/}, {_id: 0, code: 1}).sort({code: -1}).limit(1).toArray()
```

Expected: one `ORDER` row, e.g. `{ field: 'ORDER', date: '20261124', value: 7 }`,
and a highest code `OD_202611240007` with the same date and a number no
higher than `value`. If the code's date is later, or its number is higher,
fix the row before going on. Repeat with `field: 'RECEIVE'`, `code: /^RC_/`
on `receives`.

### Before deleting orphan Stock

Export the listed rows so they can be put back (untried; ids from the
report):

```bash
mongoexport --uri='<NEW_MONGO_URI>' --db=<COPY_DB_NAME> --collection=product_stocks \
  --query='{"_id": {"$in": [{"$oid": "<ID_1>"}, {"$oid": "<ID_2>"}]}}' --out=orphan-stock.json
```

Expected: `exported <N> records`, N equal to the ids given. Keep the file with
the dump.

After fixing the copy, run the report again until it exits 0, or the owner
accepts the remaining findings in writing.

## 4. Apply run

Needs `-confirm-db` with the same name as `MONGO_POS_DB_NAME`, or the command
exits 2 without touching the database:

```bash
MONGO_HOST='<NEW_MONGO_URI>' MONGO_POS_DB_NAME=<COPY_DB_NAME> \
  go run ./cmd/migrate-legacy-db -apply -confirm-db <COPY_DB_NAME>
```

Expected (exit 0): the same lines as the report run, mode `APPLY`, `HQ Branch
<24 hex characters> (created: true)`, and `changed <N>` where the report said
`would change <N>`.

If findings remain it exits 3, and says `the backfill above HAS been written`:
the backfill happened anyway, the findings still need their decision.

Re-running is safe: the second run reports `changed 0` everywhere and
`created: false`.

## 5. Start pos-api against the copy and check it

Start pos-api locally (or on a staging service) with its environment pointed at
the copy. These two pin a deployment to one database and one Client:

```
MONGO_HOST=<NEW_MONGO_URI>
MONGO_POS_DB_NAME=<COPY_DB_NAME>
CLIENT_ID=<the shop's Client id>
```

The HQ Branch already exists, so `AUTO_INIT_DEFAULT_BRANCH` is not needed; on
Cloud Run the Branch is not auto-created otherwise. A user with no Employee
record gets the HQ Branch.

Checks, all must hold:

1. **Startup log has no failed index.** Expected: no line containing
   `failed to create`. A line like `failed to create <name> index: ...
   duplicate key` means a unique index could not be built: a duplicate the
   report should have named (`DUPLICATE_SERIAL`, `DUPLICATE_SEQUENCE`). Fix the
   copy and restart.
2. **Stock is visible:** sign in as a shop user, open Stock: quantities equal
   the legacy till's.
3. **A legacy bill opens:** open an old Order by its code in the till; lines
   and payments show.
4. **A past month reports:** the sales report for a month before the dump shows
   a non-zero total, equal to the legacy report for that month.

The checks above are read-only and are the only ones run on the copy that will
go live.

5. **A test Sale lowers Stock, on a throwaway copy.** A test Sale leaves a
   permanent trace: cancelling it keeps the cancelled Order, its lines and
   payments, product history rows and the consumed order code, and today's
   counter has moved. So by default do this on a **second, throwaway
   restore of the same dump** (section 2 under another `<COPY_DB_NAME>`),
   migrated the same way (sections 3 and 4), and drop it afterwards. Sell one
   unit of a test Product; its Stock falls by one. If the owner chooses to
   test in the live-bound copy anyway, record the Order code of the test Sale.

## 6. The switch

Only after the owner's go-ahead for this step.

1. Confirm the shop has stopped selling on the legacy till (section 1).
2. [ ] **Required: compare the legacy database with the copy.** This reads the
   live legacy database. It writes nothing, but it touches the legacy stack,
   so it needs the owner's explicit go-ahead and is never run by an agent on
   its own. Untried. Run the same script against both and compare:

   ```bash
   cat > compare.js <<'JS'
   const n = db.getName();
   print(n, 'orders', db.orders.countDocuments({}));
   print(n, 'payments', db.payments.countDocuments({}));
   print(n, 'receives', db.receives.countDocuments({}));
   print(n, 'newest order createdDate', JSON.stringify(db.orders.find({}, {createdDate: 1, _id: 0}).sort({createdDate: -1}).limit(1).toArray()));
   print(n, 'sequences', JSON.stringify(db.sequences.find({}, {_id: 0, field: 1, date: 1, value: 1}).sort({field: 1}).toArray()));
   JS
   mongosh '<LEGACY_MONGO_URI>' --quiet compare.js
   mongosh '<NEW_MONGO_URI>/<COPY_DB_NAME>' --quiet compare.js
   ```

   Expected: every line identical except the database name at the start. The
   migration adds no documents to these collections, so counts match exactly.
   On any mismatch, something
   was sold or received on the legacy till after the dump: do not switch. Take
   a fresh dump (section 2), then run the report and apply again (sections 3
   and 4) and compare again.
3. Deploy or point the shop's pos-api at `<COPY_DB_NAME>` with `MONGO_HOST`,
   `MONGO_POS_DB_NAME` and `CLIENT_ID` set as in section 5, and give the shop
   the current till URL (`devper.web.app`).
4. The shop signs in and makes one real sale; check Stock fell and the bill
   printed.

The legacy services and their database are not touched at any point.

## 7. Going back

Point the shop at the old URL (`devper-pos.web.app`). The legacy services were
never modified, and their database never changed.

- Going back **before any real sale** on the new stack costs nothing.
- Going back **after** costs every bill and stock movement made on the new
  stack: they exist only in the copy, and would have to be re-keyed on the
  legacy till.
