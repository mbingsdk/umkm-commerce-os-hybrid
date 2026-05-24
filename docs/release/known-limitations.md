# Known Limitations - v1.0.0-pilot

`v1.0.0-pilot` is intentionally limited. These are product boundaries, not bugs, unless the behavior contradicts the current API or UI contract.

## Payments

- Manual payment confirmation only.
- No payment gateway integration.
- No automatic bank reconciliation.
- No payout or settlement workflow.
- No automated refund ledger.

## POS

- POS is online-first for MVP.
- No offline POS mode.
- No local device sync conflict handling.
- No receipt printer integration guarantee.

## Commerce and marketplace

- Checkout is one store per transaction.
- Discovery sends customers to tenant storefronts/products.
- No marketplace-wide checkout.
- No marketplace sync to Shopee/Tokopedia/TikTok Shop.
- No paid ads or recommendation ranking engine.

## Storefront and SEO

- No custom tenant domain support yet.
- Public storefront uses platform routes such as `/s/[storeSlug]`.
- Sitemap should include only published stores and active public products.
- Dashboard/admin/private pages must remain excluded from sitemap and robots allow rules.

## Finance

- Finance `net_estimate` is an operational estimate.
- It does not include detailed HPP/COGS calculation yet.
- It does not implement tax accounting, accrual accounting, settlement, refund ledger, or full bookkeeping.

## Courier

- Courier zones support basic local rates.
- No driver mobile app.
- No route optimization, delivery slots, COD, or proof-of-delivery workflow beyond the current shipment tracking foundation.

## Admin and plans

- Super admin tools are operational basics only.
- No billing/subscription payment gateway.
- No invoice generation.
- No custom role builder.

## AI and automation

- No AI SEO assistant yet.
- No AI recommendations.
- No external notification sending is implemented by default; outbox handlers are safe placeholders unless configured later.

## Pilot expectation

Use this release with a small pilot group, active support, regular backups, and manual review of orders, stock, and payment flows.
