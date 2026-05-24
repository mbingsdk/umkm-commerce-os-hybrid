# Pilot Go-Live Checklist

This checklist prepares the first UMKM pilot cohort for `v1.0.0-pilot`.

## 1. Pilot scope

```txt
[ ] Pilot tenant count is intentionally small.
[ ] Each pilot tenant has a named owner/contact.
[ ] Support channel is agreed before launch.
[ ] Known limitations are explained in plain Indonesian.
[ ] Manual payment workflow is accepted by pilot tenants.
[ ] Offline POS is explicitly out of scope for this pilot.
```

## 2. Data and access readiness

```txt
[ ] Super admin account exists and is secured.
[ ] Demo data is available only in local/staging, not mixed with production tenant data.
[ ] Pilot owner accounts are created or owners can self-register.
[ ] Tenant/store onboarding flow is tested.
[ ] Store publish/unpublish behavior is understood.
[ ] Courier zones are prepared if the tenant uses local delivery.
```

## 3. Tenant onboarding smoke flow

For each pilot tenant:

```txt
[ ] Owner logs in.
[ ] Owner creates or verifies store profile.
[ ] Owner creates categories.
[ ] Owner creates products with active/draft status as needed.
[ ] Product images upload correctly.
[ ] Initial stock is correct.
[ ] Store is published.
[ ] Public storefront URL is shared.
```

## 4. Customer flow smoke test

```txt
[ ] Storefront loads on mobile.
[ ] Product detail loads on mobile.
[ ] Customer can add item to cart.
[ ] Checkout creates an order.
[ ] Manual payment instruction is clear.
[ ] Payment confirmation can be submitted.
[ ] Owner can review and confirm/reject payment.
[ ] Order status can move through the intended flow.
```

## 5. POS and inventory smoke test

```txt
[ ] Cashier can open POS session.
[ ] Product search is fast enough for cashier use.
[ ] POS cash transaction succeeds when stock is enough.
[ ] POS rejects insufficient stock.
[ ] POS session can be closed with expected cash/difference.
[ ] Stock adjustment in/out works with audit trail.
```

## 6. Reports and operations

```txt
[ ] Dashboard summary loads.
[ ] Finance summary loads and is understood as an estimate.
[ ] Expense create/edit/delete works for owner/manager.
[ ] Courier zone and shipment tracking flows work if enabled.
[ ] Discovery search routes customers to tenant storefront/product pages.
```

## 7. Go-live day

```txt
[ ] Latest production backup exists.
[ ] API and frontend health checks pass.
[ ] Support owner is online.
[ ] Incident response playbook is open.
[ ] First tenant storefront URL is verified.
[ ] First real order/POS transaction is monitored closely.
```

## 8. First-week review

```txt
[ ] Collect daily blocker reports.
[ ] Review order/stock mismatches, if any.
[ ] Review POS speed and cashier usability.
[ ] Review checkout clarity and payment confirmation confusion.
[ ] Review requested features separately from bugs.
[ ] Decide continue/pause/expand pilot.
```
