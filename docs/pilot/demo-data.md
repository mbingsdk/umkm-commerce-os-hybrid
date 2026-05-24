# Demo Data - UMKM Commerce OS Hybrid

This data is for local, staging, and pilot walkthroughs only. It is not production customer data.

## How to seed

Run after migrations:

```bash
cd backend
go run ./cmd/seed
```

The seed is idempotent and safe to run more than once in local/staging. It will not create duplicate demo users, products, slugs, stock snapshots, or initial demo stock movements.

Production guard:

```txt
APP_ENV=production blocks the seed by default.
```

Only override for an approved disposable production-like environment:

```bash
DEMO_SEED_ALLOW_PRODUCTION=true go run ./cmd/seed
```

To skip seed entirely:

```bash
DEMO_SEED_ENABLED=false go run ./cmd/seed
```

To create the demo super admin:

```bash
DEMO_SEED_SUPER_ADMIN_ENABLED=true go run ./cmd/seed
```

## Demo password

Default non-production password:

```txt
demo-password-change-me
```

Override locally:

```bash
DEMO_SEED_PASSWORD="your-local-demo-password" go run ./cmd/seed
```

Do not reuse this password in production.

## Demo accounts

| Role | Email | Notes |
|---|---|---|
| Owner | `owner.demo@umkm.test` | Tenant owner for Toko Bunga Ayu |
| Staff | `staff.demo@umkm.test` | Staff account for permission smoke tests |
| Cashier | `cashier.demo@umkm.test` | POS cashier account |
| Super admin | `superadmin.demo@umkm.test` | Created only when `DEMO_SEED_SUPER_ADMIN_ENABLED=true` |

## Demo store

```txt
Store name: Toko Bunga Ayu
City: Makassar
Store slug: toko-bunga-ayu
Public storefront: /s/toko-bunga-ayu
Public platform URL example: http://localhost:3000/s/toko-bunga-ayu
```

The store is published and discoverable.

## Demo products

| Product | Slug | Category | Price | Initial stock | Discoverable |
|---|---|---|---:|---:|---|
| Bouquet Mawar Merah | `bouquet-mawar-merah` | Bouquet | Rp150.000 | 12 | Yes |
| Bouquet Money | `bouquet-money` | Bouquet | Rp250.000 | 8 | Yes |
| Hampers Wisuda | `hampers-wisuda` | Hampers | Rp185.000 | 10 | Yes |
| Bunga Matahari Mini | `bunga-matahari-mini` | Bouquet | Rp95.000 | 15 | No |

Each product gets a stock snapshot. Initial stock movements are created once with `reference_type = demo_seed`.

## Demo courier zones

| Zone | Rate |
|---|---:|
| Makassar Kota | Rp15.000 |
| Sekitar Makassar | Rp25.000 |

## Demo test flow

Suggested smoke flow:

1. Login as owner.
2. Select tenant `Toko Bunga Ayu`.
3. Open dashboard summary.
4. Review categories and products.
5. Open public storefront `/s/toko-bunga-ayu`.
6. Add `Bouquet Mawar Merah` to cart.
7. Checkout using a demo customer phone.
8. Confirm payment from dashboard.
9. Process order status.
10. Open POS with cashier account.
11. Create one POS transaction.
12. Check inventory movements.
13. Check finance summary.
14. If super admin demo is enabled, suspend and reactivate the tenant from `/admin`.

