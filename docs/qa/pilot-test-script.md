# Sprint 11F Pilot Test Script

Manual pilot QA script for local/staging verification. Use only demo data; do not paste production secrets or real customer data.

## Preconditions

- Backend, worker, PostgreSQL, and frontend are running.
- Migrations have been applied.
- Use a disposable demo owner and store, for example:
  - Owner: `owner-demo@example.test`
  - Store: `Toko Bunga Ayu`
  - City: `Makassar`
  - Store slug: `toko-bunga-ayu`

## 1. Owner register/login

1. Open `/register`.
2. Register owner with demo-only email/password.
3. Verify redirect to onboarding.
4. Logout, then login again from `/login`.

Expected:

- Login succeeds.
- Access token is stored client-side for dashboard requests.
- No sensitive token is shown in UI.

## 2. Create tenant/store

1. Complete create-store onboarding:
   - Tenant/store name: `Toko Bunga Ayu`
   - Slug: `toko-bunga-ayu`
   - WhatsApp: demo phone number
   - City: `Makassar`
2. Verify dashboard loads with active tenant selected.

Expected:

- Store exists.
- Dashboard requests include `Authorization` and `X-Tenant-ID`.

## 3. Publish store

1. Open store settings.
2. Publish the store.
3. Open `/s/toko-bunga-ayu` in an incognito/public session.

Expected:

- Public store loads without login.
- Public request does not send `Authorization` or `X-Tenant-ID`.

## 4. Create category/product

1. Create category: `Bouquet`.
2. Create product:
   - Name: `Bouquet Mawar Merah`
   - Slug: `bouquet-mawar-merah`
   - Status: active
   - Price: `150000`
   - Initial stock: `10`
   - Discoverable: enabled if plan allows
3. Create a second draft product and verify it is hidden from public pages.

Expected:

- Active product appears in dashboard and public store.
- Draft/inactive product does not appear publicly.
- Product stock snapshot exists.

## 5. Upload product image

1. Upload a JPEG/PNG/WebP image to the active product.
2. Reload product detail page.

Expected:

- Image appears publicly.
- API response uses a public URL, not a local filesystem path.

## 6. Set/adjust stock

1. Open inventory detail for the product.
2. Adjust stock in by `5`.
3. Adjust stock out by `1`.

Expected:

- Available/on-hand values update.
- Stock movements are created.
- Adjustment out cannot exceed available stock.

## 7. Public storefront browse

1. Open `/s/toko-bunga-ayu`.
2. Filter by category.
3. Search by `mawar`.
4. Open `/s/toko-bunga-ayu/products/bouquet-mawar-merah`.

Expected:

- Category/search works.
- Rupiah formatting is correct.
- Product detail does not expose `cost_price`.

## 8. Customer checkout

1. Add product to cart.
2. Open `/s/toko-bunga-ayu/cart`.
3. Continue to checkout.
4. Fill customer and shipping address.
5. Select courier zone if configured.
6. Submit once; then retry double-click behavior.

Expected:

- Checkout creates one order.
- Frontend subtotal is labelled as estimate.
- Backend response final total is used.
- Double submit does not create duplicate order for same pending submit.

## 9. Payment confirmation

1. Use public payment confirmation flow or API collection for the order.
2. Submit payer name, bank, transfer amount/date, phone verification, and optional proof image reference.

Expected:

- Public confirmation is recorded.
- Order is not auto-paid until dashboard review.

## 10. Order processing

1. Owner opens `/dashboard/orders`.
2. Open order detail.
3. Confirm payment.
4. Move status through valid transitions:
   - paid
   - processing
   - ready_to_ship
   - shipped/completed as applicable

Expected:

- Payment status becomes paid.
- Status timeline records each change.
- Invalid transitions are rejected.

## 11. Cancel flow

1. Create a new unpaid checkout order.
2. Cancel it from order detail with reason and note.

Expected:

- Reserved stock is released.
- Second cancel does not release stock twice.
- Cancelled unpaid order is not counted as online sales.

## 12. POS open session

1. Open `/dashboard/pos`.
2. Open session with opening cash `200000`.

Expected:

- Active session is shown.
- Same user cannot open a second active session for same tenant/store.

## 13. POS transaction

1. Search product in POS.
2. Add to cart.
3. Choose cash.
4. Enter amount paid >= total.
5. Submit payment.

Expected:

- Receipt appears.
- Stock snapshot decreases.
- `pos_sale` stock movement is created.
- Double submit does not create two transactions.

## 14. POS close session

1. Close POS session.
2. Enter closing cash.

Expected:

- Expected cash and difference are calculated.
- Closed session cannot be used for another transaction.

## 15. Finance summary/expense

1. Open `/dashboard/finance`.
2. Verify online/POS sales totals.
3. Add expense:
   - Category: `operasional`
   - Amount: `50000`
   - Date: today
4. Verify net estimate changes.

Expected:

- Finance summary is tenant-scoped.
- Net estimate clearly says it does not include detailed HPP/COGS.

## 16. Courier zone/shipment/tracking

1. Create courier zones:
   - `Makassar Kota` rate `15000`
   - `Pickup Sendiri` rate `0`
2. Create a shipment from paid/ready order.
3. Update shipment status to delivered.
4. Open public tracking page `/s/toko-bunga-ayu/track-order`.
5. Search by order number + customer phone.

Expected:

- Public checkout reads active courier zones.
- Shipment status logs are visible in dashboard.
- Public tracking hides internal notes and tenant IDs.
- Delivered shipment moves order to final delivered/completed state.

## 17. Discovery search

1. Open `/`, `/explore`, `/stores`, `/products`, and `/search?q=bouquet`.
2. Search by store/product/city/category.

Expected:

- Published active store appears.
- Active discoverable product appears.
- Unpublished/suspended/non-discoverable records are hidden.
- Public discovery requests do not send auth or tenant headers.

## 18. Super admin tenant suspend/activate

1. Login as demo super admin.
2. Open `/admin/tenants`.
3. Suspend the demo tenant.
4. Verify dashboard access is blocked for tenant users.
5. Verify public store/discovery hide the tenant.
6. Activate tenant again.

Expected:

- Admin endpoints do not require `X-Tenant-ID`.
- Admin mutations create admin audit logs.
- Tenant status changes affect dashboard and public visibility.

