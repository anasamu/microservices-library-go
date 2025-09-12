# GitHub Actions for Automated Versioning

This document explains how to use GitHub Actions for automated versioning and release management in this microservices library.

## Available Workflows

### 1. Release Workflow (`release.yml`)

Automatically creates releases and tags for all modules when changes are pushed to the main branch.

#### Triggered Automatically:
- When commits are pushed to `main` or `master` branch

#### Manual Trigger:
- Through GitHub Actions UI with custom version input

#### Features:
- Automatic semantic version increment based on commit messages:
  - `BREAKING CHANGE` or `breaking change` → Major version bump
  - `feat:` → Minor version bump
  - All other commits → Patch version bump
- Updates all go.mod files with consistent versioning
- Runs `go mod tidy` on all modules
- Creates Git tags for all modules:
  - Root repository: `v1.2.3`
  - Main modules: `module-name/v1.2.3`
  - Provider modules: `ai/providers/provider-name/v1.2.3`
- Creates GitHub Release with changelog

#### Manual Release Process:
1. Go to Actions tab in GitHub
2. Select "Release" workflow
3. Click "Run workflow"
4. Enter version (e.g., `v1.2.3`) or select release type
5. Click "Run workflow"

### 2. Validate Versioning Workflow (`validate-versioning.yml`)

Runs on all pull requests to validate versioning consistency.

#### Triggered Automatically:
- On all pull requests to `main` or `master` branch

#### Features:
- Checks for duplicate versions in go.mod files
- Validates against invalid pseudo-versions
- Runs `go mod tidy` to validate all modules
- Checks module consistency

## Setting Up GitHub Actions

### Prerequisites:
1. Repository must be hosted on GitHub
2. Go 1.21+ must be available in workflows
3. Proper repository permissions for tagging

### Required Secrets:
No special secrets are required for these workflows. The `GITHUB_TOKEN` is automatically provided.

## Using Semantic Versioning

### Commit Message Conventions:
To automatically determine version increments, use these commit message prefixes:

- `feat:` - Minor version bump (new features)
- `fix:` - Patch version bump (bug fixes)
- `BREAKING CHANGE:` - Major version bump (in commit body)
- `chore:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:` - Patch version bump

### Examples:
```
feat: add new AI provider support
fix: resolve authentication issue
BREAKING CHANGE: change interface for cache provider
```

## Manual Version Management

### Creating a Specific Release:
1. Navigate to Actions → Release workflow
2. Click "Run workflow"
3. Select "manual" trigger
4. Enter desired version (e.g., `v2.1.0`)
5. Run workflow

### Version Format:
Always use semantic versioning format: `vMAJOR.MINOR.PATCH`

Examples:
- `v1.0.0` - Initial release
- `v1.0.1` - Patch fix
- `v1.1.0` - New features
- `v2.0.0` - Breaking changes

## Module Tagging Structure

The release workflow creates tags for all modules following this structure:

1. **Root Repository**: `v1.2.3`
2. **Main Modules**: `module-name/v1.2.3`
   - `ai/v1.2.3`
   - `auth/v1.2.3`
   - `cache/v1.2.3`
   - etc.
3. **Provider Modules**: `ai/providers/provider-name/v1.2.3`
   - `ai/providers/anthropic/v1.2.3`
   - `ai/providers/openai/v1.2.3`
   - etc.

## Consuming Tagged Versions

External projects can consume specific versions using:

```bash
# For the entire library
go get github.com/anasamu/microservices-library-go@v1.2.3

# For specific modules
go get github.com/anasamu/microservices-library-go/ai@ai/v1.2.3
go get github.com/anasamu/microservices-library-go/auth@auth/v1.2.3

# For provider modules
go get github.com/anasamu/microservices-library-go/ai/providers/openai@ai/providers/openai/v1.2.3
```

## Troubleshooting

### Common Issues:

1. **Workflow fails due to permission issues**:
   - Ensure GitHub Actions have proper permissions for tagging
   - Check repository settings for Actions permissions

2. **Version conflicts**:
   - Ensure the version doesn't already exist as a tag
   - Delete conflicting tags if needed

3. **Module validation fails**:
   - Run `go mod tidy` locally and commit changes
   - Check for duplicate versions in go.mod files

### Debugging Steps:

1. Check workflow logs in GitHub Actions
2. Verify local go.mod files are consistent
3. Ensure all internal dependencies use the same version pattern
4. Run validation workflow locally using the provided scripts

## Best Practices

1. **Always use conventional commit messages** for automatic versioning
2. **Run validation workflow** before merging pull requests
3. **Keep versions synchronized** across all modules
4. **Test external imports** after creating new releases
5. **Document breaking changes** in commit messages
6. **Use semantic versioning** consistently
7. **Tag all modules** together for consistency

## Customization

To customize the workflows:

1. Modify `.github/workflows/release.yml` for release process changes
2. Modify `.github/workflows/validate-versioning.yml` for validation rules
3. Adjust module lists in the workflows if you add/remove modules
4. Update version increment logic based on your team's conventions

## Integration with CI/CD

These workflows can be integrated with your CI/CD pipeline:

1. **Pre-merge validation**: Validate versioning on PRs
2. **Post-merge release**: Automatically create releases on main branch
3. **Notification**: Set up Slack or email notifications for releases
4. **Deployment**: Trigger deployment workflows after successful releases

The automated versioning ensures that all modules are consistently versioned and properly tagged, making it easier for external projects to consume specific versions of the library.