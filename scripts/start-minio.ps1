param(
    [string]$MinioExe = "D:\tools\minio\minio.exe",
    [string]$DataDir = "D:\minio-data",
    [string]$Address = ":9000",
    [string]$ConsoleAddress = ":9001",
    [string]$RootUser = "noteweb",
    [string]$RootPassword = "noteweb123",
    [string]$PidFile = "D:\tmp\noteweb-minio.pid",
    [string]$LogFile = "D:\tmp\noteweb-minio.log"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $MinioExe)) {
    Write-Error "MinIO executable not found: $MinioExe. Download it from https://dl.min.io/server/minio/release/windows-amd64/minio.exe"
}

if (Test-Path -LiteralPath $PidFile) {
    $existingPid = (Get-Content -LiteralPath $PidFile -Raw).Trim()
    if ($existingPid -and (Get-Process -Id $existingPid -ErrorAction SilentlyContinue)) {
        Write-Output "MinIO is already running. PID=$existingPid"
        Write-Output "Console: http://localhost$ConsoleAddress"
        exit 0
    }
}

New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $PidFile) | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $LogFile) | Out-Null

$env:MINIO_ROOT_USER = $RootUser
$env:MINIO_ROOT_PASSWORD = $RootPassword

$arguments = @(
    "server",
    $DataDir,
    "--address",
    $Address,
    "--console-address",
    $ConsoleAddress
)

$process = Start-Process `
    -FilePath $MinioExe `
    -ArgumentList $arguments `
    -RedirectStandardOutput $LogFile `
    -RedirectStandardError $LogFile `
    -WindowStyle Hidden `
    -PassThru

Set-Content -LiteralPath $PidFile -Value $process.Id

Start-Sleep -Seconds 2

if (Get-Process -Id $process.Id -ErrorAction SilentlyContinue) {
    Write-Output "MinIO started. PID=$($process.Id)"
    Write-Output "API: http://localhost$Address"
    Write-Output "Console: http://localhost$ConsoleAddress"
    Write-Output "User: $RootUser"
    Write-Output "Log: $LogFile"
} else {
    Write-Error "MinIO failed to start. Check log: $LogFile"
}
