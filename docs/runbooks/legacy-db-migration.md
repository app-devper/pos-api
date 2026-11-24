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
- [ ] `<NEW_MONGO_URI>` is a **replica set** (Atlas counts; a `mongos` is
      accepted too). The current API writes Stock in MongoDB transactions,
      which a standalone server refuses. It runs **MongoDB 5.0 or later**: the
      findings use `$setWindowFields` (tried on 8.0 only).
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

`mongodump` and `mongorestore` were not installed where this runbook was
written, so steps 1 and 2 are untried.

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

3. Compare counts. In `mongosh '<NEW_MONGO_URI>'` (the `mongosh` line tried
   on a synthetic scratch database; the comparison with a real dump untried):

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
database <COPY_DB_NAME> on <HOST:PORT[,HOST:PORT...]>, mode REPORT
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

**Read the first line before anything else.** It names the database and the
server's host(s) from `MONGO_HOST` (never the user, password or options). If
the host is not the server that holds the copy, or the name is not
`<COPY_DB_NAME>`, stop. The command cannot tell a copy from the live legacy
database: both are a 2024 database without Branches. This line and
`-confirm-db` are the only guards. A `MONGO_HOST` that does not parse as a
`mongodb://` or `mongodb+srv://` URI is refused with exit 2, without printing
it.

The all-zero HQ id means no HQ Branch exists yet; `-apply` creates it. Each
`Change` line is a field the migration sets; in report mode the counts are
the documents that would change.

**Exit 3** means findings follow: each prints as `FINDING <KIND> (<total>):
<detail>`, then up to 20 document ids. What each kind means and what to
decide, before section 4, is in
[legacy-db-findings.md](legacy-db-findings.md). After fixing the copy, run the
report again until it exits 0, or the owner accepts the remaining findings in
writing.

## 4. Apply run

Needs `-confirm-db` with the same name as `MONGO_POS_DB_NAME`, or the command
exits 2 without touching the database:

```bash
MONGO_HOST='<NEW_MONGO_URI>' MONGO_POS_DB_NAME=<COPY_DB_NAME> \
  ./migrate-legacy-db -apply -confirm-db <COPY_DB_NAME>
```

Read the first line again: same host and database as the report run, mode
`APPLY`. Expected (exit 0): the same lines as the report run, `HQ Branch <24
hex characters> (created: true)`, and `changed <N>` where the report said
`would change <N>`. In apply mode `changed` is what the database actually
modified; on a copy nobody else writes to, the two are equal.

If findings remain it exits 3, and says `the backfill above HAS been written`:
the backfill happened anyway, the findings still need their decision.

If it exits 1 during apply (an error, a lost connection, the 10-minute limit),
the output still lists what was done before the failure (findings, HQ
Branch, each `changed` line completed), then a line starting `the run did not
finish`; the error itself is on stderr. That line says `nothing was written`
when the run stopped before its first write (and always in report mode);
otherwise the step that failed may have written part of its documents, and
its count may be missing or partial. Run it again: a re-run is safe and
finishes what was left. Every write is filtered by a missing `branchId`, and a Receive
gets its `branchId` last, so a document already done is never written twice.

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
   duplicate key` means a unique index could not be built. For
   `products.serialNumber` and `sequences.field` the report names the
   duplicates (`DUPLICATE_SERIAL`, `DUPLICATE_SEQUENCE`). `customers.code` and
   `categories.value` are unique indexes too (`app/data/repositories/
   customer.go`, `category.go`; log names `customers`, `categories`), which
   the findings do not check. The 2024 build defined the same two unique
   indexes (`b165b4d`, same files) but ignored the error when it could not
   build them, so a legacy database may hold duplicates there only if the
   index is missing from it. Fix the copy and restart.
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
from this repo; that is the owner's call, made before step 6.4.

Each step needs the owner's go-ahead; none is run by an agent on its own.

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

   Expected: `diff` prints nothing, exit 0 (the migration changes nothing the
   script counts; the HQ Branch goes to `branches`, not compared).

   Tried 2026-11-24 on synthetic scratch databases on a local test replica
   set (URI with `?replicaSet=…`), never real data: identical → no output; one
   extra history row and one lowered Stock → exactly those two lines, exit 1.

   It sees sales, Receives, Stock moves and new Products, Units, prices or
   Customers; 2024 Product, Unit, price and Stock edits write a
   `product_histories` row (unless that write failed: the 2024 code ignores
   it). It **cannot** see an edit that changes no count, no Stock quantity and
   writes no history row (a renamed Customer; suppliers, categories, settings,
   employees). Section 1's "no use of the back office" covers those.

   On any difference: do not switch. Take a fresh dump (section 2), run the
   report and apply again (sections 3, 4), compare again. **Hand fixes made in
   the old copy are gone**: the fresh report shows those findings again (keep
   a list of your fixes, to repeat them).
4. [ ] **Repoint `pos-dev-api` at the copy.** This changes the configuration
   of the service that serves the current till: owner-authorised step only,
   never run by an agent on its own. Set, as in section 5:
   `MONGO_HOST` → `<NEW_MONGO_URI>`, `MONGO_POS_DB_NAME` → `<COPY_DB_NAME>`,
   `CLIENT_ID` → the shop's Client id. `MONGO_HOST` is a secret: how it is
   switched (a new secret, or a new version of the current one) is the
   owner's decision; a new version of a secret changes it for every service
   that reads its latest version. If the copy is on a different server than
   today's, change the Mongo host and `MONGO_POS_DB_NAME` / `CLIENT_ID` in ONE
   revision, so no revision serves `<COPY_DB_NAME>` on the old host. Untried
   example for the two plain values:

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
