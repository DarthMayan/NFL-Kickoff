# Load Test Simple para Kubernetes HPA
Write-Host "Iniciando load test agresivo..." -ForegroundColor Green
Write-Host "Presiona Ctrl+C para detener" -ForegroundColor Yellow
Write-Host ""

$counter = 0
$errors = 0

while ($true) {
    try {
        $counter++
        $result = Invoke-RestMethod -Uri "http://localhost:8080/api/teams" -TimeoutSec 2

        if ($counter % 100 -eq 0) {
            Write-Host "Requests completados: $counter (Errores: $errors)" -ForegroundColor Cyan
        }
    }
    catch {
        $errors++
        if ($errors % 10 -eq 0) {
            Write-Host "Error en request #$counter - Total errores: $errors" -ForegroundColor Red
        }
    }
}
