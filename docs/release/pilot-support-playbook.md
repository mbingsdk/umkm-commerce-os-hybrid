# Pilot Support Playbook

Use this playbook during the controlled pilot. Keep the tone with tenants calm, practical, and honest: acknowledge the issue, protect their data, then fix or route it.

## 1. Triage severity

| Severity | Examples | First response |
|---|---|---|
| P0 | suspected tenant data leak, admin bypass, stock corruption across tenants, production database down | Pause affected flows, notify technical owner, preserve logs, start incident response. |
| P1 | checkout/POS unavailable, payment confirmation broken, order cannot be processed, severe stock mismatch | Assign owner immediately, collect request_id/time/order IDs, apply safe workaround. |
| P2 | UI bug, confusing copy, slow non-critical page, report mismatch | Log with reproduction steps and schedule fix. |
| P3 | feature request, cosmetic issue, training question | Add to feedback backlog. |

Never ask tenants to share passwords, tokens, or production secrets.

## 2. Bug intake template

```txt
Tenant/store:
User role:
Page/API flow:
Time and timezone:
Request ID if visible:
Order/product/transaction ID if relevant:
Steps to reproduce:
Expected result:
Actual result:
Screenshot/video:
Business impact:
```

Avoid collecting unnecessary customer personal data. If a screenshot contains phone/address/payment proof content, store it only in the approved support channel.

## 3. Tenant feedback collection

Use `docs/pilot/feedback-form.md` after onboarding and after the first real transactions. Separate feedback into:

```txt
[ ] Bug/blocker
[ ] Confusing UX/copy
[ ] Training/support need
[ ] Requested feature
[ ] Business/process mismatch
```

Good weekly pilot questions:

- Apa langkah yang paling membingungkan minggu ini?
- Apakah checkout dan konfirmasi pembayaran mudah dijelaskan ke customer?
- Apakah POS cukup cepat untuk kasir?
- Apakah stok terasa akurat setelah order/POS/adjustment?
- Laporan mana yang paling membantu keputusan harian?

## 4. Order or stock inconsistency response

When stock/order data looks wrong:

1. Do not delete records manually.
2. Capture order ID, product ID, POS transaction ID, tenant ID, and request time.
3. Take a database backup before corrective action.
4. Inspect order status, payment status, stock reservations, stock snapshot, and stock movements.
5. Prefer a documented corrective stock movement/adjustment over direct snapshot edits.
6. Check whether the issue came from checkout race, POS race, cancel release, payment confirmation, or manual adjustment.
7. After correction, run the relevant regression or QA script from `docs/qa/`.

If multiple tenants appear affected, treat it as P0 and follow `docs/ops/incident-response.md`.

## 5. Pausing a pilot tenant safely

Use pause/suspend when a tenant needs to stop taking orders while data is reviewed.

```txt
[ ] Super admin changes tenant status to suspended.
[ ] Verify tenant dashboard access is blocked.
[ ] Verify public storefront is hidden.
[ ] Verify discovery no longer lists the tenant/products.
[ ] Notify the tenant owner with a short reason and next update time.
[ ] Preserve audit logs and avoid destructive changes.
```

Do not delete tenant data during pilot support unless there is a separate, approved data-retention decision.

## 6. Communication notes

- Be specific about what is known and unknown.
- Give the next update time if the fix is not immediate.
- Do not promise payment gateway, offline POS, marketplace sync, custom domain, or advanced finance features in the pilot.
- Link known limitations from `docs/release/known-limitations.md` when helpful.
