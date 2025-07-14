# Release Process

This document describes how to create releases and make binaries available for download on GitHub.

## Automated Releases

The project uses GitHub Actions to automatically build and release binaries when tags are pushed.

### Creating a Release

1. **Update the version in README.md** (if needed):
   ```bash
   # Update the version badge and changelog
   vim README.md
   ```

2. **Commit and push changes**:
   ```bash
   git add .
   git commit -m "Prepare for v1.0.0 release"
   git push origin main
   ```

3. **Create and push a tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **GitHub Actions automatically**:
   - Builds binaries for all platforms
   - Creates a GitHub release
   - Uploads binaries as downloadable assets

### Supported Platforms

The automated build creates binaries for:

- **Linux**: AMD64, ARM64
- **macOS**: Intel (AMD64), Apple Silicon (ARM64)
- **Windows**: AMD64

## Manual Release Process

If you need to create a manual release:

### 1. Build All Platforms

```bash
make build-all
```

This creates binaries in the `build/` directory:
- `restic-backup-checker-linux-amd64`
- `restic-backup-checker-linux-arm64`
- `restic-backup-checker-darwin-amd64`
- `restic-backup-checker-darwin-arm64`
- `restic-backup-checker-windows-amd64.exe`

### 2. Create GitHub Release

1. Go to your repository on GitHub
2. Click "Releases" → "Create a new release"
3. Choose or create a tag (e.g., `v1.0.0`)
4. Fill in the release title and description
5. Upload the binaries from the `build/` directory
6. Click "Publish release"

### 3. Build with Specific Version

To build with a specific version:

```bash
VERSION=v1.0.0 make build-all
```

## Release Notes Template

Use this template for release notes:

```markdown
## Changes

### New Features
- Feature description

### Bug Fixes
- Bug fix description

### Improvements
- Improvement description

## Installation

Download the appropriate binary for your platform:

- **Linux (x64)**: `restic-backup-checker-linux-amd64`
- **Linux (ARM64)**: `restic-backup-checker-linux-arm64`
- **macOS (Intel)**: `restic-backup-checker-darwin-amd64`
- **macOS (Apple Silicon)**: `restic-backup-checker-darwin-arm64`
- **Windows (x64)**: `restic-backup-checker-windows-amd64.exe`

Make the binary executable (Linux/macOS):
```bash
chmod +x restic-backup-checker-*
```

## Quick Start

```bash
# Authenticate with OneDrive
./restic-backup-checker login

# Set up monitoring
./restic-backup-checker setup

# Start monitoring
./restic-backup-checker
```
```

## Troubleshooting

### GitHub Actions Fails

1. **Check permissions**: Ensure the repository has Actions enabled
2. **Check secrets**: The `GITHUB_TOKEN` is automatically provided
3. **Check workflow file**: Verify the YAML syntax is correct
4. **Action versions**: Ensure you're using the latest versions of GitHub Actions (v4 for artifacts)

### Build Failures

1. **Go version**: Ensure Go 1.21+ is being used
2. **Dependencies**: Run `make deps` to ensure all dependencies are available
3. **Platform-specific issues**: Check the matrix build logs for specific errors

### Upload Failures

1. **File paths**: Verify the artifact paths in the workflow
2. **Asset names**: Ensure asset names don't conflict with existing assets
3. **Release creation**: Check if the release was created successfully before uploads

## Testing the Release Process

To test the release process without creating a public release:

1. **Create a draft release**:
   - Set `draft: true` in the workflow
   - Or manually create a draft release on GitHub

2. **Test locally**:
   ```bash
   # Test the build process
   make build-all
   
   # Test version injection
   VERSION=v1.0.0-test make build
   ./build/restic-backup-checker version
   ```

3. **Test the workflow**:
   - Push a tag like `v1.0.0-test`
   - Check the Actions tab for build results
   - Verify all platforms build successfully 