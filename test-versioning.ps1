# PowerShell script to test versioning locally before pushing to GitHub

param(
    [Parameter(Mandatory=$false)]
    [string]$Version = "v0.0.0-test"
)

Write-Output "Testing versioning with version: $Version"

# Get current directory
$currentDir = Get-Location
Write-Output "Working directory: $currentDir"

# Backup current go.mod files
Write-Output "Creating backup of go.mod files..."
Get-ChildItem -Recurse -Filter "go.mod" | ForEach-Object {
    Copy-Item $_.FullName "$($_.FullName).backup" -Force
    Write-Output "Backed up $($_.Name)"
}

try {
    # Update all go.mod files with test version
    Write-Output "Updating go.mod files with test version..."
    
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
    
    # Function to update go.mod file with test version
    function Update-GoModTestVersion {
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
        
        # Update internal dependencies to use test version
        # Match patterns like: github.com/anasamu/microservices-library-go/*/types v*
        $content = $content -replace 'github.com/anasamu/microservices-library-go/([^/]+)/types v[^\s]+', "github.com/anasamu/microservices-library-go/`$1/types $Version"
        
        # Match patterns like: github.com/anasamu/microservices-library-go/ai/providers/* v*
        $content = $content -replace 'github.com/anasamu/microservices-library-go/ai/providers/([^/\s]+) v[^\s]+', "github.com/anasamu/microservices-library-go/ai/providers/`$1 $Version"
        
        # Fix any duplicate version issues
        $content = $content -replace "(github.com/anasamu/microservices-library-go/[^/\s]+/[^/\s]+) $Version $Version", "`$1 $Version"
        $content = $content -replace "(github.com/anasamu/microservices-library-go/[^/\s]+/providers/[^/\s]+) $Version $Version", "`$1 $Version"
        
        # Write the updated content back to the file
        Set-Content $goModFile $content
        Write-Output "Updated $goModFile"
    }
    
    # Update main modules
    foreach ($module in $modules) {
        Update-GoModTestVersion $module
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
            
            # Fix any duplicate version issues
            $content = $content -replace "(github.com/anasamu/microservices-library-go/ai/types) $Version $Version", "`$1 $Version"
            
            # Write the updated content back to the file
            Set-Content $goModFile $content
            Write-Output "Updated $goModFile"
        }
    }
    
    Write-Output "All go.mod files updated with test version: $Version"
    Write-Output "You can now run 'go mod tidy' in each module to validate the changes"
    
} catch {
    Write-Output "Error occurred: $_"
    Write-Output "Restoring backup files..."
    
    # Restore backup files on error
    Get-ChildItem -Recurse -Filter "go.mod.backup" | ForEach-Object {
        $originalFile = $_.FullName -replace '\.backup$', ''
        Copy-Item $_.FullName $originalFile -Force
        Write-Output "Restored $originalFile"
    }
    
    exit 1
}

Write-Output ""
Write-Output "To restore original files, run: "
Write-Output "Get-ChildItem -Recurse -Filter 'go.mod.backup' | ForEach-Object { `$originalFile = `$_`.FullName -replace '\.backup`$', ''; Copy-Item `$_`.FullName `$originalFile -Force }"