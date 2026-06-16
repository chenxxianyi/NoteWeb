param(
    [string]$PidFile = "D:\tmp\noteweb-minio.pid"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $PidFile)) {
    Write-Output "MinIO PID file not found. Nothing to stop."
    exit 0
}

$pidText = (Get-Content -LiteralPath $PidFile -Raw).Trim()
if (-not $pidText) {
    Remove-Item -LiteralPath $PidFile -Force
    Write-Output "MinIO PID file was empty. Removed it."
    exit 0
}

$process = Get-Process -Id $pidText -ErrorAction SilentlyContinue
if (-not $process) {
    Remove-Item -LiteralPath $PidFile -Force
    Write-Output "MinIO process was not running. Removed stale PID file."
    exit 0
}

Stop-Process -Id $process.Id -Force
Remove-Item -LiteralPath $PidFile -Force
Write-Output "MinIO stopped. PID=$($process.Id)"
