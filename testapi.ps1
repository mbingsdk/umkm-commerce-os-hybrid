# REGISTER
$body = @{
  name = "Owner A"
  email = "owner-a@example.com"
  phone = "08123456789"
  password = "password123"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body

# LOGIN
$loginBody = @{
  email = "owner-a@example.com"
  password = "password123"
} | ConvertTo-Json

$login = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body $loginBody

$token = $login.data.access_token

# CEK SESSION
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/me" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $token" }

# Create tenant + store:
$storeBody = @{
  tenant_name = "Toko Bunga Ayu"
  tenant_slug = "toko-bunga-ayu"
  store = @{
    name = "Toko Bunga Ayu"
    slug = "toko-bunga-ayu"
    description = "Toko bunga lokal untuk bouquet dan hampers."
    whatsapp = "08123456789"
    city = "Makassar"
    province = "Sulawesi Selatan"
  }
} | ConvertTo-Json -Depth 5

$created = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/onboarding/create-store" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $token" } `
  -Body $storeBody

$created
$tenantId = $created.data.tenant.id

# LIST Tenant
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/tenants" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $token" }

# Cek current store dengan X-Tenant-ID:
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/stores/current" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }

# Test publish:
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/stores/current/publish" `
  -Method POST `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }

# CHECKOUT
$checkoutBody = @{
  items = @(
    @{
      product_id = "a75eb8a3-cabf-41d5-a0e5-00f27120df99"
      quantity = 1
    }
  )
  customer = @{
    name = "Customer Demo"
    phone = "08123456789"
  }
  shipping_address = @{
    recipient_name = "Customer Demo"
    recipient_phone = "08123456789"
    address = "Jl. Pengujian No. 1"
    city = "Makassar"
    province = "Sulawesi Selatan"
  }
  payment_method = "manual_transfer"
} | ConvertTo-Json -Depth 6

$key = [guid]::NewGuid().ToString()

$res1 = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/public/stores/toko-bunga-ayu/checkout" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{ "Idempotency-Key" = $key } `
  -Body $checkoutBody

$res1

# GET ORDER DETAIL + CANCEL ORDER
$orderId = $res1.data.order_id
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/orders/$orderId" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }
  
  $cancelBody = @{
  reason = "customer_request"
  note = "Customer minta dibatalkan."
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/orders/$orderId/cancel" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $cancelBody

# CEK STATUS ORDER
$res2 = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/orders/$orderId" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }
$res2

# TEST INVENTORY - STOCKS
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/inventory/stocks" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }

# TEST INVENTORY - ADJUSTMENT
$adjustBody = @{
  adjustment_type = "in"
  quantity = 5
  reason = "stock_opname"
  note = "Tambah stok dari hasil opname."
} | ConvertTo-Json

$productId = "a75eb8a3-cabf-41d5-a0e5-00f27120df99"
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/inventory/products/$productId/adjust" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $adjustBody

# TEST POS - OPEN SESSION
$openBody = @{
  opening_cash_amount = 200000
  note = "Buka kasir pagi"
} | ConvertTo-Json

$session = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/pos/sessions/open" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $openBody

$sessionId = $session.data.id

# TEST POS - SEARCH PRODUCTS
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/pos/products?q=bouquet" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }

# TEST POS - CREATE TRANSACTION
$posBody = @{
  session_id = $sessionId
  items = @(
    @{
      product_id = $productId
      quantity = 1
    }
  )
  payment_method = "cash"
  amount_paid = 100000
  note = "Transaksi test POS"
} | ConvertTo-Json -Depth 6

$key = [guid]::NewGuid().ToString()

$trx = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/pos/transactions" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
    "Idempotency-Key" = $key
  } `
  -Body $posBody

$trx

# TEST POS - CLOSE SESSION
$closeBody = @{
  closing_cash_amount = 300000
  note = "Tutup kasir"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/pos/sessions/$sessionId/close" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $closeBody

# TEST FINANCE - SUMMARY
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/finance/summary" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }

# TEST FINANCE - CREATE EXPENSE
$expenseBody = @{
  title = "Beli pita satin"
  amount = 75000
  expense_date = "2026-05-20"
  payment_method = "cash"
  note = "Bahan bouquet"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/finance/expenses" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $expenseBody

# TEST FINANCE - SUMMARY DASHBOARD
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/dashboard/summary" `
  -Method GET `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  }