# PowerShell script to restore go.mod files from backups

Write-Output "Restoring go.mod files from backups..."

# Counter for restored files
$restoredCount = 0

# Find all backup files and restore them
Get-ChildItem -Recurse -Filter "go.mod.backup" | ForEach-Object {
    $backupFile = $_.FullName
    $originalFile = $backupFile -replace '\.backup$', ''
    
    if (Test-Path $backupFile) {
        Copy-Item $backupFile $originalFile -Force
        Write-Output "Restored $originalFile"
        $restoredCount++
    }
}

Write-Output "Restored $restoredCount go.mod files from backups."

# Remove backup files
$removedCount = 0
Get-ChildItem -Recurse -Filter "go.mod.backup" | ForEach-Object {
    Remove-Item $_.FullName -Force
    $removedCount++
}

Write-Output "Removed $removedCount backup files."
Write-Output "Restore complete!"