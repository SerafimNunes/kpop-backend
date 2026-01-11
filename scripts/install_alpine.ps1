$dest = "static/vendor/alpine.min.js"
$dir = Split-Path $dest -Parent
if (-Not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
$uri = "https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"
Write-Host "Baixando Alpine.js de $uri para $dest"
try {
    Invoke-WebRequest -Uri $uri -OutFile $dest -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ Alpine baixado para: $dest"
    # Abrir o studio com token padrão para facilitar testes
    $studioUrl = "http://localhost:8080/studio.html?token=K-LENS-DEMO"
    Write-Host "Abrindo o Studio: $studioUrl"
    try { Start-Process $studioUrl } catch {}
} catch {
    Write-Host "❌ Falha ao baixar Alpine: $_"
    Write-Host "Tente baixar manualmente e salvar em static/vendor/alpine.min.js"
}
