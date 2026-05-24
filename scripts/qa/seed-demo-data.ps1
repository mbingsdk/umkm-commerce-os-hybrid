param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$OwnerPassword = "demo-password-change-me",
  [int]$OpeningCashAmount = 200000,
  [switch]$AllowNonLocal
)

$ErrorActionPreference = "Stop"

if (-not $AllowNonLocal -and $BaseUrl -notmatch "^https?://(localhost|127\.0\.0\.1)(:\d+)?$") {
  throw "Refusing to seed non-local API '$BaseUrl'. Pass -AllowNonLocal only for disposable staging."
}

function Invoke-JsonApi {
  param(
    [Parameter(Mandatory = $true)][string]$Method,
    [Parameter(Mandatory = $true)][string]$Path,
    [hashtable]$Headers = @{},
    [object]$Body = $null
  )

  $params = @{
    Uri     = "$BaseUrl$Path"
    Method  = $Method
    Headers = $Headers
  }

  if ($null -ne $Body) {
    $params.ContentType = "application/json"
    $params.Body = ($Body | ConvertTo-Json -Depth 12)
  }

  Invoke-RestMethod @params
}

$suffix = Get-Date -Format "yyyyMMddHHmmss"
$ownerEmail = "owner-demo-$suffix@example.test"
$tenantSlug = "toko-bunga-ayu-$suffix"
$storeSlug = $tenantSlug
$customerPhone = "081234567890"

Write-Host "Creating demo owner $ownerEmail against $BaseUrl"

$register = Invoke-JsonApi -Method "POST" -Path "/api/v1/auth/register" -Body @{
  name     = "Owner Demo"
  email    = $ownerEmail
  password = $OwnerPassword
  phone    = $customerPhone
}

$token = $register.data.access_token
if (-not $token) {
  throw "Register response did not include access_token."
}

$authHeaders = @{ Authorization = "Bearer $token" }

$tenant = Invoke-JsonApi -Method "POST" -Path "/api/v1/onboarding/create-store" -Headers $authHeaders -Body @{
  tenant_name = "Toko Bunga Ayu"
  tenant_slug = $tenantSlug
  store       = @{
    name        = "Toko Bunga Ayu"
    slug        = $storeSlug
    description = "Demo toko bunga non-produksi di Makassar."
    whatsapp    = $customerPhone
    city        = "Makassar"
    province    = "Sulawesi Selatan"
  }
}

$tenantId = $tenant.data.id
if (-not $tenantId) {
  throw "Onboarding response did not include tenant id."
}

$tenantHeaders = @{
  Authorization = "Bearer $token"
  "X-Tenant-ID" = $tenantId
}

Invoke-JsonApi -Method "POST" -Path "/api/v1/stores/current/publish" -Headers $tenantHeaders | Out-Null

$category = Invoke-JsonApi -Method "POST" -Path "/api/v1/categories" -Headers $tenantHeaders -Body @{
  name        = "Bouquet"
  slug        = "bouquet"
  description = "Kategori demo bouquet."
  is_active   = $true
  sort_order  = 1
}

$categoryId = $category.data.id

$products = @(
  @{
    name                = "Bouquet Mawar Merah"
    slug                = "bouquet-mawar-merah"
    description         = "Produk demo non-produksi."
    category_id         = $categoryId
    sku                 = "BQT-MWR-DEMO"
    status              = "active"
    price               = 150000
    compare_at_price    = $null
    cost_price          = 0
    track_inventory     = $true
    initial_stock       = 10
    low_stock_threshold = 2
    is_visible          = $true
    is_discoverable     = $true
  },
  @{
    name                = "Bouquet Sunflower Ceria"
    slug                = "bouquet-sunflower-ceria"
    description         = "Produk demo non-produksi."
    category_id         = $categoryId
    sku                 = "BQT-SUN-DEMO"
    status              = "active"
    price               = 125000
    compare_at_price    = $null
    cost_price          = 0
    track_inventory     = $true
    initial_stock       = 8
    low_stock_threshold = 2
    is_visible          = $true
    is_discoverable     = $true
  }
)

$createdProducts = @()
foreach ($product in $products) {
  $createdProducts += Invoke-JsonApi -Method "POST" -Path "/api/v1/products" -Headers $tenantHeaders -Body $product
}

$zoneA = Invoke-JsonApi -Method "POST" -Path "/api/v1/courier/zones" -Headers $tenantHeaders -Body @{
  name        = "Makassar Kota"
  description = "Zona demo dalam kota Makassar."
  rate        = 15000
  is_active   = $true
  sort_order  = 1
}

Invoke-JsonApi -Method "POST" -Path "/api/v1/courier/zones" -Headers $tenantHeaders -Body @{
  name        = "Pickup Sendiri"
  description = "Customer mengambil pesanan ke toko."
  rate        = 0
  is_active   = $true
  sort_order  = 2
} | Out-Null

$session = Invoke-JsonApi -Method "POST" -Path "/api/v1/pos/sessions/open" -Headers $tenantHeaders -Body @{
  opening_cash_amount = $OpeningCashAmount
  note                = "Demo cashier session dari seed-demo-data.ps1"
}

[PSCustomObject]@{
  baseUrl       = $BaseUrl
  ownerEmail    = $ownerEmail
  ownerPassword = $OwnerPassword
  tenantId      = $tenantId
  storeSlug     = $storeSlug
  categoryId    = $categoryId
  productIds    = $createdProducts.data.id -join ","
  courierZoneId = $zoneA.data.id
  posSessionId  = $session.data.id
  storefrontUrl = "http://localhost:3000/s/$storeSlug"
} | Format-List

Write-Host "Demo seed complete. Data is non-production and safe to discard."
