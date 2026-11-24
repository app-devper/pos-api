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
`<LEGACY_DB_NAME>` (the legacy database's name on that server),
`<NEW_MONGO_URI>` (a replica set that holds the copy), `<COPY_DB_NAME>` (the
copy's name, never the legacy name), `<SHOP_CLIENT_ID>` (the shop's Client
id in UM), `<GCP_PROJECT>` / `<REGION>` (where `pos-dev-api` runs),
`<RECORDED_DB_NAME>` / `<RECORDED_CLIENT_ID>` (what step 6.1 recorded). Keep real values out of this file, shell
history you share, and commits.

## 1. What must be true first

- [ ] The owner has given the go-ahead for each step below that touches the
      legacy stack (step 2.1, the dump; step 6.3, the comparison) and for each step that reads or changes `pos-dev-api`
      (steps 6.1 and 6.4, and section 7).
- [ ] The shop's users exist in UM under the shop's Client, with System `POS`.
      Authentication goes through UM; a user missing there cannot sign in.
- [ ] The current till (`devper.web.app`) is deployed. The 2024 till cannot sell
      against the new API: a Sale without `saleId` is refused with 400.
- [ ] `<NEW_MONGO_URI>` is a **replica set** (Atlas counts). The current API
      writes Stock in MongoDB transactions, which a standalone server refuses.
- [ ] **No use of the legacy till or its back office** from the dump until the
      switch: no sale, no Receive, no Stock count or adjustment, no new or
      edited Product, Unit, price or Customer. Order codes continue from the
      `sequences` collection of the copy; a bill made on the legacy till after
      the dump is lost, and the next code can clash with it. Any other change
      made there after the dump is lost too.

## 2. Dump the legacy database and restore it as a copy

Run once the shop has stopped using the legacy till and its back office
(section 1, last item). This reads the legacy database; it writes nothing to
it, but it touches the legacy stack: it needs the owner's explicit go-ahead
and is never run by an agent on its own.

1. Dump (untried):

   ```bash
   mongodump --uri='<LEGACY_MONGO_URI>' --db=<LEGACY_DB_NAME> --out=./legacy-dump
   ```

   Expected: one `done dumping <LEGACY_DB_NAME>.<collection> (N documents)`
   line per collection, no `error` lines, exit 0.

2. Restore under a new name on the replica set (untried):

   ```bash
   mongorestore --uri='<NEW_MONGO_URI>' \
     --nsFrom='<LEGACY_DB_NAME>.*' --nsTo='<COPY_DB_NAME>.*' ./legacy-dump
   ```

   Expected: `N document(s) restored successfully. 0 document(s) failed to
   restore.`, exit 0.

3. Compare counts (untried). In `mongosh '<NEW_MONGO_URI>'`:

   ```js
   db.getSiblingDB('<COPY_DB_NAME>').orders.countDocuments({})
   ```

   Expected: the same N the dump printed for `<LEGACY_DB_NAME>.orders` (this
   reads the copy only, not the legacy database). Do the same for
   `product_stocks` and `payments`.

## 3. Report run (writes nothing)

Build the command once and run the binary. Do not use `go run`: it turns
every non-zero exit into 1 (printing `exit status 3`), which hides the exit
codes below.

```bash
go build -o migrate-legacy-db ./cmd/migrate-legacy-db
MONGO_HOST='<NEW_MONGO_URI>' MONGO_POS_DB_NAME=<COPY_DB_NAME> ./migrate-legacy-db
```

Tried 2026-11-24 on synthetic scratch databases on a local test replica set,
never real data: report and apply exit 0 clean, 3 with a finding; a re-run
reports `changed 0` and `created: false`.

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
| `DISCOUNT_PER_LINE` | An Order with a discounted line whose total is Σ(price − discount): the discount taken once per Line. | **Do NOT edit totals in the copy: they are what the shop charged.** Read "Discounted 2024 Orders after the move" below. |
| `DISCOUNT_NOT_SUBTRACTED` | An Order with a discounted line whose total is Σ price: the discount not taken at all. This is how the 2024 server itself recomputed a total. | **Do NOT edit totals in the copy: they are what the shop charged.** Read "Discounted 2024 Orders after the move" below. |
| `DISCOUNT_UNKNOWN` | An Order with a discounted line whose total matches none of the three rules. | **Do NOT edit totals in the copy: they are what the shop charged.** Open the listed Orders with the owner to understand them; read "Discounted 2024 Orders after the move" below. |
| `UNKNOWN_PAYMENT_TYPE` | `payments.type` outside `CASH`, `CREDIT`, `PROMPTPAY`, `TRANSFER`. | First find out what each legacy value meant (ask the shop); only then map it, or leave it. Effect of leaving it: the current Go code stores and returns `payments.type` without grouping or validating it, so the API itself is unaffected; whether the till or its reports group by it is not determined. |

### Discounted 2024 Orders after the move

A Line's `price` is its whole amount before discount. The current code charges
Σ(price − discount × quantity): the discount is per unit
(`app/domain/sale/money.go`, `OrderMoney`). An Order whose total follows that
rule is fine, whatever else it also matches. The check compares the stored
total of every Order with a discounted line against three rules, within 0.005:

- per unit, Σ(price − discount × quantity): fine, not reported;
- per line, Σ(price − discount): `DISCOUNT_PER_LINE`;
- not subtracted, Σ price: `DISCOUNT_NOT_SUBTRACTED`;
- none of them: `DISCOUNT_UNKNOWN`.

Where the totals came from: the 2024 server stored the total the till sent
when the bill was made, and replaced it with Σ price (discount not subtracted)
whenever a Line was deleted from the bill (`b165b4d`
`app/data/repositories/order.go`, `UpdateTotalOrderById` /
`GetTotalOrderById`). What the 2024 **till** sent as `price` and `discount`,
and which rule it used for the total, is **not known from this repo**.

The stored totals stay as they are. What changes after the move:

- **Cancelling one Line** of such an Order (`CancelLine`,
  `app/data/ledger/cancel.go`, `recomputeOrder`) recomputes the **whole**
  Order from the Lines still standing with the per-unit rule
  (`sale.OrderMoney`) and overwrites its `total`, `totalCost` and `discount`.
  The new total is then lower than what the remaining Lines were charged:
  for an Order listed under `DISCOUNT_NOT_SUBTRACTED`, by each remaining
  Line's discount × quantity; under `DISCOUNT_PER_LINE`, by each remaining
  Line's discount × (quantity − 1).
- **Cancelling the whole Order** (`CancelOrder`) marks it, its Lines and its
  payments cancelled and puts the Stock back; it does not recompute the total.
- **A Return** caps each refund at price ÷ quantity − discount per unit
  (`app/data/ledger/return.go`, `paidPerUnit`), the per-unit rule again.

So for the Orders listed under these three kinds the shop should cancel the
whole Order rather than one Line, or accept the recomputed total. Tell the
shop before the switch, with the Order ids from the report.

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
  ./migrate-legacy-db -apply -confirm-db <COPY_DB_NAME>
```

Expected (exit 0): the same lines as the report run, mode `APPLY`, `HQ Branch
<24 hex characters> (created: true)`, and `changed <N>` where the report said
`would change <N>`.

If findings remain it exits 3, and says `the backfill above HAS been written`:
the backfill happened anyway, the findings still need their decision.

If it exits 1 during apply (an error, a lost connection, the 10-minute limit),
run it again: a re-run is safe and finishes what was left. Every write is
filtered by a missing `branchId`, and a Receive gets its `branchId` last, so a
document already done is never written twice.

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
   test in the live-bound copy anyway, record the Order code of the test Sale;
   the comparison in step 6.3 will then differ by that Sale (an Order, its
   Lines, its payment, history rows, Stock and the `ORDER` counter).

## 6. The switch

**The service that is repointed.** The current till (`devper.web.app`) reaches
pos-api through the gateway `https://api.devper.app`, route `/api/pos/**`, which
goes to one Cloud Run service: `pos-dev-api`. That service serves one database
and one Client, pinned by `MONGO_POS_DB_NAME` and `CLIENT_ID`; its `MONGO_HOST`
comes from Secret Manager (a `pos-api-*` secret). Its configuration lives on the
service and survives deploys (AGENTS.md, "Environment Notes").

**Repointing `pos-dev-api` at the copy takes whatever it serves today off the
current till.** Whether anyone depends on what it serves today is **not known**
from this repo; that is the owner's call, made before step 4.

Each step needs the owner's go-ahead for that step, and none is run by an agent
on its own.

1. [ ] **Record what `pos-dev-api` serves today, before any change.** Reads the
   service's configuration only. Untried example (`<GCP_PROJECT>`, `<REGION>`
   are the service's project and region):

   ```bash
   gcloud run services describe pos-dev-api --project=<GCP_PROJECT> --region=<REGION> \
     --format='yaml(status.latestReadyRevisionName,spec.template.spec.containers[0].env)'
   ```

   Expected: the revision name and an env list with `MONGO_POS_DB_NAME`,
   `CLIENT_ID`, and `MONGO_HOST` as a secret reference (name and version, not
   the value). Write down those, the revision, and from the owner which Client
   and data that database holds and who uses it today. Keep this record with
   the dump, not in a commit; section 7 restores these values.
2. Confirm no use of the legacy till or its back office since the dump.
3. [ ] **Required: compare the legacy database with the copy.** This reads the
   live legacy database. It writes nothing, but it touches the legacy stack,
   so it needs the owner's explicit go-ahead and is never run by an agent on
   its own. From the repo root; each side is given its database name:

   ```bash
   S=docs/runbooks/legacy-db-compare.js
   COMPARE_DB=<LEGACY_DB_NAME> mongosh '<LEGACY_MONGO_URI>' --quiet $S > legacy.txt
   COMPARE_DB=<COPY_DB_NAME> mongosh '<NEW_MONGO_URI>' --quiet $S > copy.txt
   diff legacy.txt copy.txt
   ```

   Expected: `diff` prints nothing, exit 0. The migration adds no documents to
   these collections and changes no Stock quantity or counter (the HQ Branch
   goes to `branches`, not compared), so every line matches exactly.

   What was tried (2026-11-24, never real data): this script file on two
   synthetic scratch databases on a local test replica set, URI with
   `?replicaSet=…`. Identical databases: no `diff` output. One extra
   `product_histories` row and one lowered Stock quantity: exactly those two
   lines, `diff` exit 1. No `COMPARE_DB`: a message, exit 2.

   What it sees: a sale, a Receive, a Stock added, adjusted or counted, a new
   Product, Unit, price or Customer. In the 2024 build every Product, Unit,
   price and Stock edit also writes a `product_histories` row, so an edit of
   those is seen through that count (the 2024 code ignores a failed history
   write, so such an edit could in principle leave no row). What it **cannot**
   see: an edit that changes no count, no Stock quantity and writes no
   history row, for example a renamed Customer, a changed Customer type, or
   anything in a collection it does not count (suppliers, categories,
   settings, employees). Section 1's "no use of the back office" is what
   covers those.

   On any difference: do not switch. Something was done on the legacy stack
   after the dump. Take a fresh dump (section 2), then run the report and
   apply again (sections 3 and 4) and compare again. **Hand fixes made in the
   old copy are gone** with it: the report run on the fresh copy shows those
   findings again, and each needs its fix again (keep a list of the fixes you
   made, so they can be repeated).
4. [ ] **Repoint `pos-dev-api` at the copy.** This changes the configuration
   of the service that serves the current till: owner-authorised step only,
   never run by an agent on its own. Set, as in section 5:
   `MONGO_HOST` → `<NEW_MONGO_URI>`, `MONGO_POS_DB_NAME` → `<COPY_DB_NAME>`,
   `CLIENT_ID` → the shop's Client id. `MONGO_HOST` is a secret: how it is
   switched (a new secret, or a new version of the current one) is the
   owner's decision; a new version of a secret changes it for every service
   that reads its latest version. Untried example for the two plain values:

   ```bash
   gcloud run services update pos-dev-api --project=<GCP_PROJECT> --region=<REGION> \
     --update-env-vars=MONGO_POS_DB_NAME=<COPY_DB_NAME>,CLIENT_ID=<SHOP_CLIENT_ID>
   ```

   Expected: a new revision ready, serving 100% of traffic.
5. Give the shop the current till URL (`devper.web.app`). The shop signs in
   and makes one real sale; check Stock fell and the bill printed.

The legacy services and their database are not modified at any point.

## 7. Going back

Point the shop at the old URL (`devper-pos.web.app`). The legacy services were
never modified, and their database never changed.

Then restore `pos-dev-api` to the values recorded in step 6.1: `MONGO_HOST`
(the recorded secret and version), `MONGO_POS_DB_NAME` and `CLIENT_ID`, so it
serves again what it served before the switch. This changes that service's
configuration: owner-authorised step only, never run by an agent on its own.
Untried example:

```bash
gcloud run services update pos-dev-api --project=<GCP_PROJECT> --region=<REGION> \
  --update-env-vars=MONGO_POS_DB_NAME=<RECORDED_DB_NAME>,CLIENT_ID=<RECORDED_CLIENT_ID>
```

Expected: a new revision ready with the recorded values.

- Going back **before any real sale** on the new stack costs nothing.
- Going back **after** costs every bill and stock movement made on the new
  stack: they exist only in the copy, and would have to be re-keyed on the
  legacy till.
