@echo off
REM Build script for CRM Backend

echo Cleaning up old dependencies...
go clean -modcache

echo Downloading dependencies...
go mod download

echo Tidying up...
go mod tidy

echo Building...
go build -o crm-backend.exe ./cmd/server

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo You can run the server with: crm-backend.exe
) else (
    echo Build failed. Please check the errors above.
)
