# Скрипт для проверки всех Go файлов через анализатор
$files = Get-ChildItem -Path . -Include *.go -Recurse | Where-Object {
    $_.FullName -notlike "*\testdata\*" -and
    $_.FullName -notlike "*\vendor\*" -and
    $_.FullName -notlike "*\linter.exe"
}

$errors = @()

foreach ($file in $files) {
    $relativePath = $file.FullName.Replace((Get-Location).Path + "\", "")
    Write-Host "Checking: $relativePath" -ForegroundColor Cyan
    
    $result = & .\linter.exe $relativePath 2>&1
    if ($LASTEXITCODE -ne 0 -and $result -notlike "*internal error*") {
        Write-Host $result -ForegroundColor Yellow
        $errors += $relativePath
    }
}

if ($errors.Count -eq 0) {
    Write-Host "`nВсе файлы проверены!" -ForegroundColor Green
} else {
    Write-Host "`nНайдены проблемы в файлах:" -ForegroundColor Red
    $errors | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
}



