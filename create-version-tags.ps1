# PowerShell script to create version tags for all modules
# This script helps with transitioning from pseudo-versions to semantic versioning

param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

# Validate version format (should be v1.2.3 format)
if (-not ($Version -match '^v\d+\.\d+\.\d+')) {
    Write-Error "Invalid version format. Please use format: v1.2.3"
    exit 1
}

Write-Output "Creating version tags for all modules with version: $Version"

# List of all main modules
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

# Create tag for root repository
Write-Output "Creating tag for root repository..."
git tag -a "$Version" -m "Release $Version"

# Create tags for each module
foreach ($module in $modules) {
    if (Test-Path $module) {
        $moduleTag = "$module/$Version"
        Write-Output "Creating tag: $moduleTag"
        git tag -a "$moduleTag" -m "Release $module $Version"
    } else {
        Write-Output "Warning: Directory $module not found, skipping..."
    }
}

# Create tags for AI providers
$providers = @("anthropic", "deepseek", "google", "openai", "xai")
foreach ($provider in $providers) {
    $providerPath = "ai\providers\$provider"
    if (Test-Path $providerPath) {
        $providerTag = "ai/providers/$provider/$Version"
        Write-Output "Creating tag: $providerTag"
        git tag -a "$providerTag" -m "Release ai/providers/$provider $Version"
    }
}

Write-Output "All version tags created successfully!"
Write-Output "To push tags to remote repository, run: git push --tags"