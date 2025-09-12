# PowerShell script to run go mod tidy on all modules

Write-Output "Running go mod tidy on all modules..."

# List of all modules
$modules = @(
    "ai"
    "auth" 
    "backup"
    "cache"
    "chaos"
    "circuitbreaker"
    "communication"
    "config"
    "database"
    "discovery"
    "event"
    "failover"
    "filegen"
    "logging"
    "messaging"
    "middleware"
    "monitoring"
    "payment"
    "ratelimit"
    "scheduling"
    "storage"
)

# Run go mod tidy on each module
foreach ($module in $modules) {
    if (Test-Path $module) {
        Write-Output "Running go mod tidy in $module..."
        Set-Location $module
        go mod tidy
        Set-Location ..
        Write-Output "Completed $module"
    } else {
        Write-Output "Warning: Directory $module not found, skipping..."
    }
}

Write-Output "All modules have been tidied!"