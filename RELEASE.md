# Release Process

This document describes how to create a release of jotr.

## Prerequisites

- Go 1.21+
- Git
- GitHub CLI (optional but recommended)

## Building Release Binaries

### Using Make

```bash
# Build for all platforms
make build-all

# This creates binaries in dist/:
# - jotr-darwin-amd64
# - jotr-darwin-arm64
# - jotr-linux-amd64
# - jotr-linux-arm64
# - jotr-windows-amd64.exe
```

### Manual Build

```bash
# Create dist directory
mkdir -p dist

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dist/jotr-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dist/jotr-darwin-arm64

# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o dist/jotr-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o dist/jotr-linux-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/jotr-windows-amd64.exe
```

## Creating a Release

### 1. Update Version

Update version in:
- `main.go` (if hardcoded)
- `README.md`
- `jotr.rb` (Homebrew formula)

### 2. Create Git Tag

```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag
git push origin v1.0.0
```

### 3. Build Binaries

```bash
make build-all
```

### 4. Create GitHub Release

#### Using GitHub CLI

```bash
# Create release with binaries
gh release create v1.0.0 \
  dist/jotr-darwin-amd64 \
  dist/jotr-darwin-arm64 \
  dist/jotr-linux-amd64 \
  dist/jotr-linux-arm64 \
  dist/jotr-windows-amd64.exe \
  --title "jotr v1.0.0" \
  --notes "Release notes here"
```

#### Using GitHub Web UI

1. Go to https://github.com/yourusername/jotr/releases/new
2. Choose tag: `v1.0.0`
3. Release title: `jotr v1.0.0`
4. Add release notes
5. Upload binaries from `dist/`
6. Publish release

### 5. Update Homebrew Formula

Update `jotr.rb` with:
- New version number
- New download URLs
- New SHA256 checksums

```bash
# Calculate SHA256
shasum -a 256 dist/jotr-darwin-amd64
shasum -a 256 dist/jotr-darwin-arm64
shasum -a 256 dist/jotr-linux-amd64
shasum -a 256 dist/jotr-linux-arm64
```

Update the formula and commit:

```bash
git add jotr.rb
git commit -m "Update Homebrew formula to v1.0.0"
git push
```

## Automated Release (GitHub Actions)

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build binaries
        run: make build-all
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            dist/jotr-darwin-amd64
            dist/jotr-darwin-arm64
            dist/jotr-linux-amd64
            dist/jotr-linux-arm64
            dist/jotr-windows-amd64.exe
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Then just push a tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will automatically build and create the release!

## Release Checklist

- [ ] Update version numbers
- [ ] Update CHANGELOG.md
- [ ] Run tests: `make test`
- [ ] Build all binaries: `make build-all`
- [ ] Test binaries on each platform
- [ ] Create git tag
- [ ] Push tag
- [ ] Create GitHub release
- [ ] Upload binaries
- [ ] Update Homebrew formula
- [ ] Update documentation
- [ ] Announce release

## Testing Release

Before publishing, test each binary:

```bash
# macOS Intel
./dist/jotr-darwin-amd64 version

# macOS Apple Silicon
./dist/jotr-darwin-arm64 version

# Linux
./dist/jotr-linux-amd64 version

# Windows (on Windows or Wine)
./dist/jotr-windows-amd64.exe version
```

## Post-Release

1. Update documentation with new version
2. Announce on social media / blog
3. Update any package managers (Homebrew, etc.)
4. Close milestone on GitHub
5. Start planning next release!

