---
status: accepted
---

# Sold first stays separate from Oversell

A Line the cashier did not mark for Oversell, sold when no Stock covers it, is not rejected: the uncovered quantity goes to the Product's Sold first, the shop's unreconciled bucket. We kept that behaviour rather than rejecting the sale (the till does not block it, so cashiers would meet a new refusal at checkout) or folding Sold first into Oversell debt (Oversell is a deliberate per-Line promise to a waiting customer; Sold first is quantity nobody accounted for). Sold first is written only by the Stock ledger, is never settled by incoming Stock, and is cleared by hand.
