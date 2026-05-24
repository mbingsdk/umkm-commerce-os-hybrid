# Pilot Onboarding Guide

This guide helps onboard the first UMKM pilot users without improvising the flow each time.

## Before inviting a pilot UMKM

Prepare:

- Staging or pilot production URL.
- Demo walkthrough account if needed.
- WhatsApp/support contact.
- Short expectation: this is MVP pilot, feedback is expected.
- Backup and rollback process confirmed.

Suggested invitation copy:

```txt
Halo, kami sedang membuka pilot terbatas UMKM Commerce OS Hybrid.
Tujuannya membantu toko mengelola produk, order, stok, POS kasir, laporan dasar, dan storefront online.
Selama pilot, kami akan mendampingi setup awal dan mencatat feedback dari penggunaan nyata.
```

## Owner creates store

1. Owner registers account.
2. Owner logs in.
3. Owner completes create-store onboarding.
4. Confirm store identity:
   - store name
   - city
   - WhatsApp number
   - address
5. Publish store when ready.

Checklist:

```txt
[ ] Owner can login.
[ ] Tenant/store created.
[ ] Store profile complete enough for customer.
[ ] Store published.
[ ] Public store URL opens.
```

## Add first products

Start with 3-5 real products.

Recommended fields:

- Product name.
- Product slug.
- Category.
- Description.
- Price in Rupiah.
- Stock quantity.
- Product photo.
- Active/draft status.
- Discoverable if store wants platform listing.

Checklist:

```txt
[ ] Category created.
[ ] Product draft created.
[ ] Product active created.
[ ] Initial stock visible.
[ ] Public product detail opens.
```

## Process first online order

1. Customer opens public storefront.
2. Customer adds one product to cart.
3. Customer fills checkout form.
4. Owner sees order in dashboard.
5. Customer submits manual payment confirmation.
6. Owner confirms payment.
7. Owner moves order through processing/ready/shipped/completed.

Checklist:

```txt
[ ] Checkout succeeds.
[ ] Order appears in dashboard.
[ ] Stock is reserved.
[ ] Payment confirmation can be reviewed.
[ ] Order status can be updated.
[ ] Stock movement is recorded.
```

## Use POS

1. Cashier logs in.
2. Cashier opens POS session with opening cash.
3. Cashier searches product.
4. Cashier adds product to cart.
5. Cashier enters cash paid.
6. Cashier completes transaction.
7. Cashier checks receipt dialog.
8. Cashier closes session at end of shift.

Checklist:

```txt
[ ] POS session opens.
[ ] Product search is fast enough.
[ ] POS transaction succeeds.
[ ] Stock decreases.
[ ] Cash change is clear.
[ ] Session close shows expected cash and difference.
```

## Check reports

Owner should review:

- Dashboard summary.
- Recent orders.
- Low stock.
- Finance summary.
- Expenses.
- Daily/monthly reports.

Explain clearly:

```txt
Net estimate is not full accounting profit yet because detailed HPP/modal is not included.
```

Checklist:

```txt
[ ] Owner understands sales source split.
[ ] Owner can add expense.
[ ] Owner understands net estimate limitation.
[ ] Low stock card helps operational decision.
```

## Feedback session

After the first real test order or POS sale, ask the pilot user to fill `docs/pilot/feedback-form.md`.

Focus on:

- setup difficulty
- product management clarity
- checkout clarity
- POS speed
- order flow
- report usefulness
- blockers
- must-have requested features

