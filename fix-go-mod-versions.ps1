# PowerShell script to fix pseudo-versions and ensure consistent module versioning
# across all libraries in the microservices-library-go repository

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

# Function to update go.mod file with proper pseudo-version
function Update-GoModPseudoVersion {
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
    
    # Update internal dependencies to use proper pseudo-version
    # Match patterns like: github.com/anasamu/microservices-library-go/*/types v*
    $content = $content -replace 'github.com/anasamu/microservices-library-go/([^/]+)/types v[^\s]+', "github.com/anasamu/microservices-library-go/`$1/types $pseudoVersion"
    
    # Match patterns like: github.com/anasamu/microservices-library-go/ai/providers/* v*
    $content = $content -replace 'github.com/anasamu/microservices-library-go/ai/providers/([^/\s]+) v[^\s]+', "github.com/anasamu/microservices-library-go/ai/providers/`$1 $pseudoVersion"
    
    # Fix any duplicate version issues
    $content = $content -replace "(github.com/anasamu/microservices-library-go/[^/\s]+/[^/\s]+) $pseudoVersion $pseudoVersion", "`$1 $pseudoVersion"
    $content = $content -replace "(github.com/anasamu/microservices-library-go/[^/\s]+/providers/[^/\s]+) $pseudoVersion $pseudoVersion", "`$1 $pseudoVersion"
    
    # Write the updated content back to the file
    Set-Content $goModFile $content
    Write-Output "Updated $goModFile"
}

# Update main modules
foreach ($module in $modules) {
    Update-GoModPseudoVersion $module
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
        $content = $content -replace 'github.com/anasamu/microservices-library-go/ai/types v[^\s]+', "github.com/anasamu/microservices-library-go/ai/types $pseudoVersion"
        
        # Fix any duplicate version issues
        $content = $content -replace "(github.com/anasamu/microservices-library-go/ai/types) $pseudoVersion $pseudoVersion", "`$1 $pseudoVersion"
        
        # Write the updated content back to the file
        Set-Content $goModFile $content
        Write-Output "Updated $goModFile"
    }
}

Write-Output "All pseudo-versions updated!"