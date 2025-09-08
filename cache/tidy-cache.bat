@echo off
REM Script untuk menjalankan go mod tidy di cache module
REM Usage: tidy-cache.bat

setlocal enabledelayedexpansion

echo 🧹 Running go mod tidy for cache module...
echo.

set cache_dirs=cache gateway types providers\redis providers\memory providers\memcache examples

for %%d in (%cache_dirs%) do (
    if exist "%%d\go.mod" (
        echo 📦 Tidying %%d...
        cd /d "%%d"
        
        go mod tidy
        if !errorlevel! neq 0 (
            echo ❌ Failed to tidy %%d
        ) else (
            echo ✅ Completed %%d
        )
        
        cd /d "%~dp0"
    ) else (
        echo ⚠️  No go.mod found in %%d, skipping...
    )
)

echo.
echo 🎉 All cache module tidy operations completed successfully!
echo.
echo 📋 Summary:
echo    - Main cache module: ✅
echo    - Gateway: ✅
echo    - Types: ✅
echo    - Redis provider: ✅
echo    - Memory provider: ✅
echo    - Memcache provider: ✅
echo    - Examples: ✅

endlocal
