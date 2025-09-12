# Module Versioning Guidelines

This document explains how to maintain consistent versioning across all modules in this microservices library.

## Current Versioning Approach

All modules in this repository use pseudo-versions based on the git commit hash to maintain consistency. The format is:

```
v0.0.0-{timestamp}-{commit-hash}
```

Where:
- `timestamp` is in UTC format: `YYYYMMDDHHMMSS`
- `commit-hash` is the first 12 characters of the current git commit hash

## GitHub Actions for Automated Versioning

This repository includes GitHub Actions workflows for automated versioning:

1. **Release Workflow** - Automatically creates releases and tags for all modules
2. **Validate Versioning Workflow** - Validates versioning consistency on pull requests

See [GITHUB_ACTIONS.md](GITHUB_ACTIONS.md) for detailed information.

## Scripts for Version Management

### 1. fix-go-mod-versions.ps1 (PowerShell)

This PowerShell script updates all go.mod files to use consistent pseudo-versions:

```powershell
# Run from the repository root
powershell -ExecutionPolicy Bypass -File "fix-go-mod-versions.ps1"
```

### 2. fix-pseudo-versions.sh (Bash)

This bash script updates all go.mod files to use consistent pseudo-versions:

```bash
# Run from the repository root
./fix-pseudo-versions.sh
```

### 3. prepare-for-semantic-versioning.ps1 (PowerShell)

This script prepares the repository for semantic versioning and creates appropriate tags:

```powershell
# Run from the repository root with a specific version
powershell -ExecutionPolicy Bypass -File "prepare-for-semantic-versioning.ps1" -Version "v1.0.0"
```

### 4. test-versioning.ps1 (PowerShell)

This script tests versioning locally before pushing to GitHub:

```powershell
# Run from the repository root with a test version
powershell -ExecutionPolicy Bypass -File "test-versioning.ps1" -Version "v0.0.0-test"
```

### 5. restore-go-mod.ps1 (PowerShell)

This script restores go.mod files from backups:

```powershell
# Run from the repository root
powershell -ExecutionPolicy Bypass -File "restore-go-mod.ps1"
```

## Manual Process

If you need to manually update the versions:

1. Get the current commit hash:
   ```bash
   git rev-parse HEAD
   ```

2. Create a pseudo-version with timestamp:
   ```bash
   # Format: v0.0.0-{timestamp}-{12-char-commit-hash}
   # Example: v0.0.0-20250913054541-f9d256adf8dd
   ```

3. Update all go.mod files to use this pseudo-version for internal dependencies.

## After Version Updates

After updating versions, run `go mod tidy` in each module directory to ensure consistency:

```bash
# For each module directory
cd module-name
go mod tidy
```

Or use the provided scripts:
- `tidy-all-modules.ps1` (PowerShell)
- `tidy-all-modules.sh` (Bash)

## Versioning Strategy

For future releases, consider moving to semantic versioning:
- v1.0.0 for stable releases
- v1.0.1 for patches
- v1.1.0 for minor features
- v2.0.0 for breaking changes

This would require tagging commits appropriately with git tags.

## Troubleshooting Common Issues

### 1. Invalid Version Errors

If you see errors like:
```
invalid version: unknown revision 000000000000
```

This means some go.mod files still reference invalid pseudo-versions. Run the fix scripts to update them.

### 2. Duplicate Version Errors

If you see errors about duplicate versions in go.mod files, the fix scripts now handle this automatically.

### 3. Module Not Found Errors

If you see errors like:
```
module github.com/anasamu/microservices-library-go/filegen@latest found (v1.0.0), but does not contain package github.com/anasamu/microservices-library-go/filegen
```

This indicates that the module path doesn't match the actual package structure. Ensure that:
1. The module declaration in go.mod matches the repository path
2. The package is properly tagged in the repository
3. The go.mod file exists in the correct directory

## Best Practices

1. Always run the version fix scripts after making changes to the repository structure
2. Run `go mod tidy` in each module after version updates
3. Test external imports after making version changes
4. Use semantic versioning for stable releases
5. Keep all internal dependencies synchronized to the same version
6. Use GitHub Actions for automated versioning and validation
7. Follow conventional commit messages for automatic version incrementing