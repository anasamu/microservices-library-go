# PowerShell script to prepare the repository for semantic versioning
# This script helps transition from pseudo-versions to proper semantic versions

param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

# Validate version format (should be v1.2.3 format)
if (-not ($Version -match '^v\d+\.\d+\.\d+')) {
    Write-Error "Invalid version format. Please use format: v1.2.3"
    exit 1
}

Write-Output "Preparing repository for semantic versioning: $Version"

# Get the current commit hash
$commitHash = git rev-parse HEAD
Write-Output "Using commit hash: $commitHash"

# Create pseudo-version from commit hash (12 characters)
$shortCommitHash = $commitHash.Substring(0, 12)
$timestamp = Get-Date -UFormat "%Y%m%d%H%M%S"
$pseudoVersion = "v0.0.0-$timestamp-$shortCommitHash"
Write-Output "Using pseudo-version: $pseudoVersion"

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

# Function to update go.mod file with semantic version
function Update-GoModSemanticVersion {
    param(
        [string]$module
    )
    
    $goModFile = "$module\go.mod"
    
    if (-not (Test-Path $goModFile)) {
        Write-Output "Warning: $goModFile not found, skipping..."
        return
    }
    
    Write-Output "Processing $goModFile..."
    
    # Read the content of the go.mod file
    $content = Get-Content $goModFile -Raw
    
    # Update internal dependencies to use semantic version
    # Match patterns like: github.com/anasamu/microservices-library-go/*/types v0.0.0-*
    $content = $content -replace 'github.com/anasamu/microservices-library-go/([^/]+)/types v[^\s]+', "github.com/anasamu/microservices-library-go/`$1/types $Version"
    
    # Match patterns like: github.com/anasamu/microservices-library-go/ai/providers/* v0.0.0-*
    $content = $content -replace 'github.com/anasamu/microservices-library-go/ai/providers/([^/\s]+) v[^\s]+', "github.com/anasamu/microservices-library-go/ai/providers/`$1 $Version"
    
    # Write the updated content back to the file
    Set-Content $goModFile $content
    Write-Output "Updated $goModFile"
}

# Update main modules
foreach ($module in $modules) {
    Update-GoModSemanticVersion $module
}

# Update AI provider modules
$providers = @("anthropic", "deepseek", "google", "openai", "xai")
foreach ($provider in $providers) {
    $goModFile = "ai\providers\$provider\go.mod"
    if (Test-Path $goModFile) {
        Write-Output "Processing $goModFile..."
        
        # Read the content of the go.mod file
        $content = Get-Content $goModFile -Raw
        
        # Update ai/types dependency
        $content = $content -replace 'github.com/anasamu/microservices-library-go/ai/types v[^\s]+', "github.com/anasamu/microservices-library-go/ai/types $Version"
        
        # Write the updated content back to the file
        Set-Content $goModFile $content
        Write-Output "Updated $goModFile"
    }
}

# Create tags for all modules
Write-Output "Creating version tags for all modules with version: $Version"

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
foreach ($provider in $providers) {
    $providerPath = "ai\providers\$provider"
    if (Test-Path $providerPath) {
        $providerTag = "ai/providers/$provider/$Version"
        Write-Output "Creating tag: $providerTag"
        git tag -a "$providerTag" -m "Release ai/providers/$provider $Version"
    }
}

Write-Output "Repository prepared for semantic versioning with version: $Version"
Write-Output "To push tags to remote repository, run: git push --tags"