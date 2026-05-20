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
$token = $res.data.access_token

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
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/orders/6c5ee079-872a-451c-aad6-37307a2c0804" `
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
  -Uri "http://localhost:8080/api/v1/orders/6c5ee079-872a-451c-aad6-37307a2c0804/cancel" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
    "X-Tenant-ID" = $tenantId
  } `
  -Body $cancelBody