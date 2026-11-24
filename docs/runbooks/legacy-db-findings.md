# Findings of the legacy database migration

Part of [legacy-db-migration.md](legacy-db-migration.md), section 3. Same
hard rules and placeholders: everything here runs on the restored **COPY**
(`<NEW_MONGO_URI>`, `<COPY_DB_NAME>`), never on the legacy database, and no
step is run by an agent without the owner's go-ahead for that step.

`migrate-legacy-db` exits 3 when findings follow. Each prints as
`FINDING <KIND> (<total>): <detail>`, then up to 20 document ids (smallest
first). `(<total>)` is the real count. Decide each before section 4 of the
main runbook:

| Kind | Meaning | Decision |
|---|---|---|
| `NOT_REPLICA_SET` | The server is neither a replica set nor a `mongos`. | Stop. Restore onto a replica set (main runbook, section 2). |
| `NEGATIVE_STOCK` | Stock below zero (`product_stocks`). The current ledger refuses a move that leaves one. | Do not hand-edit `quantity`: that bypasses the Stock ledger (ADR-0001). First record, per listed row, the legacy value and the counted value. Then either (a) leave it and do a Stock count or adjustment in the new till after the move, or (b) the owner decides on a hand edit in the copy. Limitation if left: the ledger refuses a move that leaves a Stock below zero, so cancelling a bill that drew from that Stock can fail until it is corrected. |
| `DUPLICATE_SERIAL` | Products share a `serialNumber`, empty counts, soft-deleted too; a Product whose `serialNumber` is an array is listed here too. The unique index on `products.serialNumber` cannot be built. | Give each a distinct serial (a plain text, not an array) in the copy. |
| `DUPLICATE_SEQUENCE` | `sequences` rows share a `field`; a row whose `field` is an array is listed here too. The unique index cannot be built. | See "Duplicate sequence rows" below; the right row is not always the highest value. |
| `ORPHAN_STOCK` | Stock whose `productId` or `unitId` matches no Product or Unit. | Export the listed rows first (see "Before deleting orphan Stock" below). Deleting them is the owner's decision; the alternative is to restore the missing Product/Unit. |
| `NO_MAIN_UNIT` | A Product has no `product_units` row named like its own unit; a Receive is entered in the main unit. | Add the unit row in the copy. |
| `DISCOUNT_PER_LINE` | An Order with a discounted line whose total is Σ(price − discount): the discount taken once per Line. | **Do NOT edit totals in the copy: they are what the shop charged.** Read "Discounted 2024 Orders after the move" below. |
| `DISCOUNT_NOT_SUBTRACTED` | An Order with a discounted line whose total is Σ price: the discount not taken at all. This is how the 2024 server itself recomputed a total. | **Do NOT edit totals in the copy: they are what the shop charged.** Read "Discounted 2024 Orders after the move" below. |
| `DISCOUNT_UNKNOWN` | An Order with a discounted line whose total matches none of the three rules. | **Do NOT edit totals in the copy: they are what the shop charged.** Open the listed Orders with the owner to understand them; read "Discounted 2024 Orders after the move" below. |
| `NON_NUMERIC_LINE_AMOUNT` | An Order (the ids are **Order** ids) with a Line whose `price`, `quantity` or `discount` the current code cannot read: not a number (such as the text `"5"`, counted as 0 by the discount check), a Decimal128, or a fractional `quantity`. | Read "Lines with an amount the current code cannot read" below. |
| `UNKNOWN_PAYMENT_TYPE` | `payments.type` outside `CASH`, `CREDIT`, `PROMPTPAY`, `TRANSFER`. | First find out what each legacy value meant (ask the shop); only then map it, or leave it. Effect of leaving it: the current Go code stores and returns `payments.type` without grouping or validating it, so the API itself is unaffected; whether the till or its reports group by it is not determined. |

The findings do **not** check `customers.code` or `categories.value`, which
are unique indexes too; the startup check (main runbook, section 5) shows them.

## Discounted 2024 Orders after the move

A Line's `price` is its whole amount before discount. The current code charges
Σ(price − discount × quantity): the discount is per unit
(`app/domain/sale/money.go`, `OrderMoney`). An Order whose total follows that
rule is fine, whatever else it also matches. The check compares the stored
total of every Order with a discounted line against three rules, within 0.005:

- per unit, Σ(price − discount × quantity): fine, not reported;
- per line, Σ(price − discount): `DISCOUNT_PER_LINE`;
- not subtracted, Σ price: `DISCOUNT_NOT_SUBTRACTED`;
- none of them: `DISCOUNT_UNKNOWN`.

The check sums the raw line amounts, while the current code rounds each Line
to the satang first, so a bill whose line amounts have fractions of a satang
may be listed as `DISCOUNT_UNKNOWN` although it is fine.

The 2024 server stored the total the till sent, and replaced it with Σ price
whenever a Line was deleted (`b165b4d` `app/data/repositories/order.go`,
`UpdateTotalOrderById`). What the 2024 **till** sent as `price` and
`discount`, and its rule for the total, is **not known from this repo**.

The stored totals stay as they are. What changes after the move:

- **Cancelling one Line** (`CancelLine`, `app/data/ledger/cancel.go`,
  `recomputeOrder`) recomputes the **whole** Order from its standing Lines
  with the per-unit rule (`sale.OrderMoney`), overwriting `total`,
  `totalCost`, `discount`. The new total is lower than what the remaining
  Lines were charged: by each one's discount × quantity under
  `DISCOUNT_NOT_SUBTRACTED`, × (quantity − 1) under `DISCOUNT_PER_LINE`.
- **Cancelling the whole Order** (`CancelOrder`) marks it, its Lines and its
  payments cancelled and puts the Stock back; it does not recompute the total.
- **A Return** caps each refund at price ÷ quantity − discount per unit
  (`app/data/ledger/return.go`, `paidPerUnit`). On a `DISCOUNT_NOT_SUBTRACTED`
  or `DISCOUNT_PER_LINE` bill (per line: quantity above 1) that is below what
  was paid per unit: the shop decides the refund amount by hand.

The report lists at most 20 ids per kind. So the rule for the shop does not
depend on that list: **for any bill made before the switch that has a
discounted Line, cancel the whole bill, never one Line.** To give the shop
every such bill, run on the COPY (reads only; tried on a synthetic scratch
database only, never real data):

```bash
COPY_DB=<COPY_DB_NAME> mongosh '<NEW_MONGO_URI>' --quiet docs/runbooks/legacy-discounted-orders.js
```

Expected: one line per Order, `<_id> <code> <createdDate>`, oldest first. A
Line whose `discount` is text is not listed (MongoDB does not compare text
with a number); it is a `NON_NUMERIC_LINE_AMOUNT` bill.

## Lines with an amount the current code cannot read

`NON_NUMERIC_LINE_AMOUNT` lists Orders (an id `(missing)` stands for Lines
with no `orderId`; a text `orderId` shows in quotes). To see which Line and
value, for one listed Order id, in `mongosh '<NEW_MONGO_URI>'` (reads only;
tried on a synthetic scratch database only, never real data):

```js
const d = db.getSiblingDB('<COPY_DB_NAME>');
d.order_items.aggregate([
  {$match: {orderId: ObjectId('<ORDER_ID>')}},
  {$project: {price: 1, quantity: 1, discount: 1,
    types: {price: {$type: '$price'}, quantity: {$type: '$quantity'}, discount: {$type: '$discount'}}}},
]).forEach(l => print(l._id.toHexString(), 'price', l.types.price, String(l.price),
  'quantity', l.types.quantity, String(l.quantity), 'discount', l.types.discount, String(l.discount)));
```

Expected: one line per Line of that Order, each amount with its BSON type,
for example `… price int 200 quantity long 2 discount string 5`. Fine:
`int`, `long`, `double` (for `quantity` only a whole `double` such as `2`),
and `missing` or `null` (read as 0). Not fine: `string`, `decimal`, any
other type, and a fractional, infinite or NaN `quantity` such as `2.5`.

What the current code does with such a Line: `entities.OrderItem` holds
`price` and `discount` as `float64` and `quantity` as `int`. Checked with
throwaway decode tests against the driver in `go.mod` (v1.17.9, not
committed):
- text in `price` or `discount` fails with `cannot decode string into a
  float32 or float64 type`;
- text in `quantity` fails with `cannot decode string into an integer type`;
- a Decimal128 in any of the three fails (`cannot decode 128-bit decimal`);
- a `quantity` of `2.5` or NaN fails (`float64 can only be truncated … when
  truncation is enabled`), and +Inf fails (`overflows int64`);
- a whole double `quantity`, a `long` or an `int` decodes, and so does a
  missing or null amount (as 0).

The finding lists all the failing cases. So **the bill cannot be opened,
cancelled or returned against until the value is corrected**, and a list that
reads that Line fails as a whole. The discount check counts a non-number as 0
and a Decimal128 as its value, so such an Order may also appear under a
discount kind.

Decision: look at the listed documents with the owner. Correct the value in
the copy (as a number) only if the owner can say what it was, for example
from the printed bill; otherwise leave it, and accept that the bill cannot be
opened in the new till. Record each correction with your other fixes: a fresh
dump brings the old value back.

## Duplicate sequence rows

`app/data/repositories/sequence.go`: each `sequences` row has `field`,
`value`, `date` (`YYYYMMDD`, shop time zone), `prefix`, `format` (4) and
`type`. A code is `prefix` + the date part of its type + `value` zero-padded
to 4 digits (`app/data/entities/sequence.go`, `GenerateCode`). On each
request the counter adds 1 to `value` while the row's period is the current
one; when a new period starts it sets `value` to 1 and `date` to today.

| `field` | `type` | Code, kept in | Row to keep |
|---|---|---|---|
| `ORDER` | `DAILY` | `OD_` + `YYYYMMDD` + value, `orders.code` | latest `date`, then the highest `value` |
| `RECEIVE` | `DAILY` | `RC_` + `YYYYMMDD` + value, `receives.code` | latest `date`, then the highest `value` |
| `MEMBER` | `YEARLY` | `MB_` + `YYYY` + value, `customers.code` | latest year (the first 4 characters of `date`), then the highest `value` |
| `PRODUCT` | `NONE` | `PD_` + value, `products.serialNumber` (the serial the till suggests) | the highest `value`; `date` plays no part |
| `BRANCH`, `STOCK_TRANSFER`, `STOCK_COUNT`, `STOCK_ADJUSTMENT`, `PRODUCT_RETURN` | `NONE` | value, with `TF-`, `SC-`, `AJ-`, `RT-` before it for the last four | the highest `value`. Only the current API writes these rows; the 2024 build used `ORDER`, `RECEIVE`, `MEMBER` and `PRODUCT` only, so a 2024 copy should have none. |

For a dated counter, keeping an older period with a higher value makes the
next code restart at 1 and collide with codes already used in the current
period. For every counter, **the surviving row must cover the highest code
already used**: same period as that code (dated types) and a `value` no
lower than its number. A row that does not is a clash waiting for the next
bill, Receive, member or product. Check on the COPY (reads only; tried on a
synthetic scratch database only, never real data):

```bash
COPY_DB=<COPY_DB_NAME> mongosh '<NEW_MONGO_URI>' --quiet docs/runbooks/legacy-sequence-check.js
```

Expected: one `row <_id> <field> <type> date <date> value <value>` line per
row, then per counter `<FIELD> highest code <code> date <date part> number
<n>` (or `none`). The script compares codes by date part, then by number, so
it stays right past 9,999 codes in a period (a 5-digit `OD_2026112410000`
ranks above `OD_202611249999`, which a text sort would get wrong). Numbers are
read as decimals, so a code of up to 34 digits compares right; tried with a
20-digit one. Example:
`row … ORDER DAILY date 20261124 value 7` and `ORDER highest code
OD_202611240007 date 20261124 number 7` agree. If the code's period is later
than the row's, or its number is higher, correct the surviving row in the
copy before going on.

## Before deleting orphan Stock

Export the listed rows so they can be put back (untried: `mongoexport` is not
installed where this runbook was written; ids from the report):

```bash
mongoexport --uri='<NEW_MONGO_URI>' --db=<COPY_DB_NAME> --collection=product_stocks \
  --query='{"_id": {"$in": [{"$oid": "<ID_1>"}, {"$oid": "<ID_2>"}]}}' --out=orphan-stock.json
```

Expected: `exported <N> records`, N equal to the ids given. Keep the file with
the dump.

After fixing the copy, run the report again until it exits 0, or the owner
accepts the remaining findings in writing.
