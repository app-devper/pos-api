---
status: accepted
---

# Oversell is settled from every inflow of the same Unit, drawing from it

Oversold quantity was settled two incompatible ways — a Receive import drew it from the new Lot after its transaction had committed, while a positive Adjustment left the quantity alone and recorded a synthetic `ADJUST:` marker — and both looked up debts by Product and branch only, so a delivery of tablets paid off a debt in boxes and the Line could then no longer be returned. We decided that an oversold Line is owed goods the customer is still waiting for: whenever Stock of a Unit rises — Receive import, Transfer approved into the branch, a positive Adjustment or Count, a manually created Stock, a Return, a cancelled Order or Line, a rejected Transfer — the ledger first settles that branch's outstanding debts for the same Product **and Unit**, oldest Line first, drawing the settled quantity out of the incoming Stock and recording it on the Line. So the quantity a Stock shows is what is actually free to sell, and `stock ≈ Σ remaining − Σ oversold` holds after every event.

## Considered Options

- **Settle without drawing** (the Adjustment behaviour) — rejected: it treats the oversold goods as having left the shelf at the sale, so Stock reads high until the next Count, and it breaks the invariant above.
- **Settle across Units by converting quantities** — rejected: Units have no conversion factor in the model and each Unit keeps its own Stocks; adding conversion is a separate change to what a Stock is.
- **Settle only from Receive and Transfer** — rejected: goods coming back from a Return or a cancelled bill reach the shelf just the same, and every exception is a path where the invariant can drift.

## Consequences

- A Line's cost is fixed when it is sold; settlement does not rewrite it, so recorded profit never changes after the fact.
- A debt lives on its Line until settled; nothing ties it to a particular Stock in the meantime.
- Cancelling an Order or Line puts its quantity back into every Stock it drew from, settlement draws included; its own unsettled debt disappears.
- Debts already settled from the wrong Unit are repaired by `cmd/repair-cross-unit-oversell` (report by default, `-apply` to write). Past settlements that did not draw are left alone: drawing them now would contradict Counts taken since.
