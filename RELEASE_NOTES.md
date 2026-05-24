# Release Notes - v1.0.0-pilot

Release date: 2026-05-25

## Summary

`v1.0.0-pilot` is the first pilot-ready release candidate for UMKM Commerce OS Hybrid. It is intended for controlled staging and limited UMKM pilot usage, not broad public production rollout.

The release focuses on:

- multi-tenant storefront and dashboard foundation
- auth, tenant onboarding, role permissions, and tenant isolation
- catalog, inventory, product images, and public storefront
- cart, checkout, order management, manual payment confirmation, and stock reservation
- POS session and POS transaction flows
- finance summary and basic expense tracking
- courier zones, shipment tracking, and public discovery
- super admin tenant/plan/featured discovery management
- security hardening, QA scripts, deployment assets, backup/restore docs, and release checklists

## Pilot-ready areas

```txt
[x] Owner can create and publish a store.
[x] Tenant can manage categories/products and initial stock.
[x] Public storefront and product detail pages are SEO-friendly.
[x] Customer can checkout from one store.
[x] Order and stock reservation flow exists.
[x] Manual payment confirmation exists.
[x] POS online-first flow exists.
[x] Basic finance summary and expenses exist.
[x] Courier zones and shipment tracking exist.
[x] Discovery pages route customers to tenant storefronts/products.
[x] Super admin can manage tenants/plans/featured items.
[x] Deployment, backup, restore, QA, and support docs exist.
```

## Known limitations

This pilot release intentionally does not include:

- payment gateway integration
- automatic bank/payment reconciliation
- offline POS
- marketplace sync
- custom domain support
- advanced accounting ledger
- detailed HPP/COGS profit calculation
- tax/refund/payout automation
- AI SEO assistant
- native mobile apps

See `docs/release/known-limitations.md` for the full pilot explanation.

## Required pre-release checks

Before tagging/deploying:

```txt
[ ] Complete RELEASE_CHECKLIST.md.
[ ] Complete staging deployment checklist.
[ ] Complete production deployment checklist if going beyond staging.
[ ] Complete pilot go-live checklist before inviting UMKM.
```

## Operational notes

- Use manual payment confirmation during pilot.
- Keep pilot tenant count small.
- Take database backup before every deploy.
- Test restore on staging before production restore.
- Treat tenant data leak, stock corruption, checkout/POS race bug, and admin bypass as high severity.

