---
status: accepted
---

# The Stock ledger is the only way a Stock's quantity changes

Stock was moved by seven repository files with three transaction settings and five separate ways of writing product history, so the rules that keep the shelf honest (never below zero, one history row with its balance, Oversell settled) were each enforced in some paths and not others. We decided that one module, `app/data/ledger`, is the only code that changes `product_stocks.quantity`, creates or deletes a Stock, writes product history, or changes a Line's `stocks[]` / `oversoldQty` — every Sale draw, Order and Line cancel, Receive import, Transfer request / approve / reject, Adjustment, Count, Return, set-quantity, and manual Stock create and delete. Each call is one transaction (snapshot read, majority write) that writes the change, exactly one history row per Stock touched with its resulting balance, and any Oversell settlement. The ledger also owns every document status whose repetition would move Stock twice (Receive IMPORTED, Transfer APPROVED / REJECTED, Line CANCELLED), so it is the one place that refuses a repeat.

## Considered Options

- **One generic `Record(event)` entry point with an in-transaction hook** — rejected: one result type for every event and a hook with re-run rules put the interface back on callers.
- **An open `Movement` interface that callers compose inside a `Run`** — rejected for now: callers would have to know settlement flags, history labels and retry rules, which is the knowledge the ledger exists to take away. Its closed set of steps (take, give, open, close) is kept *inside* the ledger as the single choke point.
- **One method per domain event (chosen)** — `Sell`, `ImportReceive`, `Adjust`, `Count`, `Return`, `CancelOrder`, `CancelLine`, `RequestTransfer`, `ApproveTransfer`, `RejectTransfer`, `CreateStock`, `DeleteStock`, `SetQuantity`, each taking the request its handler already has and returning the document it already returns. Callers distinguish three outcomes: rejected (with a Thai reason), not found, and conflict (safe to resend).

## Consequences

- Product history names a Product and Unit, not a Stock. So a Sale or a cancel writes one history row per Line, as it always has, even when the Line drew from several Stocks or from none (Sold first, Oversell); every other event writes one row per Stock it touched.

- `Sell` owns recording a Sale end to end — the Order, its Lines and payments, the `saleId` repeat check, and the Order code, taken inside the transaction so a rejected Sale no longer burns a code. The repeat check identifies a Sale by its Lines and Customer, not its payments, so a retry with a different tender returns the recorded Order.
- Rollout is three PRs: (1) the ledger with Receive import, Adjustment, Count and Return; (2) Sale and cancel; (3) Transfer, manual Stock create / delete and set-quantity. After (3), `app/data/ledger/guard_test.go` fails any write outside the ledger that changes a Stock's quantity, creates or deletes a Stock, moves a Line's `stocks[]` / `oversoldQty`, or moves Sold first. The only exceptions it allows are listed there with their reason: a new Product's opening Stock (no Line can owe it), clearing Sold first by hand (ADR-0003), and the legacy till-priced Order path below.
- The legacy till-priced Order path (`CreateOrder`, requests without `saleId`) is not moved into the ledger; it is deleted once production logs show no till still uses it.
- Tests run against a real Mongo replica set through the ledger's methods. There is no in-memory ledger: with one adapter, that seam would be hypothetical.
