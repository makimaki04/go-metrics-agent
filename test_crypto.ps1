# Тестовый скрипт для проверки шифрования

Write-Host "=== Тест асимметричного шифрования ===" -ForegroundColor Green
Write-Host ""

# Проверяем наличие ключей
if (-not (Test-Path "private_key.pem")) {
    Write-Host "ОШИБКА: private_key.pem не найден!" -ForegroundColor Red
    Write-Host "Запустите: go run generate_keys.go" -ForegroundColor Yellow
    exit 1
}

if (-not (Test-Path "public_key.pem")) {
    Write-Host "ОШИБКА: public_key.pem не найден!" -ForegroundColor Red
    Write-Host "Запустите: go run generate_keys.go" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Ключи найдены" -ForegroundColor Green
Write-Host ""

# Проверяем компиляцию
Write-Host "Компиляция сервера..." -ForegroundColor Cyan
go build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "ОШИБКА компиляции сервера!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Сервер скомпилирован" -ForegroundColor Green

Write-Host "Компиляция агента..." -ForegroundColor Cyan
go build -o agent.exe ./cmd/agent
if ($LASTEXITCODE -ne 0) {
    Write-Host "ОШИБКА компиляции агента!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Агент скомпилирован" -ForegroundColor Green
Write-Host ""

Write-Host "=== Инструкция по тестированию ===" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Запустите сервер в отдельном терминале:" -ForegroundColor White
Write-Host "   .\server.exe -a :8080 -crypto-key private_key.pem" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. Запустите агент в другом терминале:" -ForegroundColor White
Write-Host "   .\agent.exe -a localhost:8080 -crypto-key public_key.pem" -ForegroundColor Cyan
Write-Host ""
Write-Host "3. Проверьте логи сервера - должны быть сообщения о:" -ForegroundColor White
Write-Host "   - 'Private key loaded successfully'" -ForegroundColor Green
Write-Host "   - Успешной обработке запросов" -ForegroundColor Green
Write-Host ""
Write-Host "4. Проверьте логи агента - должны быть сообщения о:" -ForegroundColor White
Write-Host "   - Успешной отправке метрик" -ForegroundColor Green
Write-Host ""
Write-Host "5. Проверьте без шифрования (обратная совместимость):" -ForegroundColor White
Write-Host "   Сервер: .\server.exe -a :8080" -ForegroundColor Cyan
Write-Host "   Агент:  .\agent.exe -a localhost:8080" -ForegroundColor Cyan
Write-Host ""
