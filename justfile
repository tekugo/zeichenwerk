# Build all packages
build:
    go build ./...

# Run all tests
test:
    go test ./...

# Run tests with coverage report
cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Run go vet
vet:
    go vet ./...

# Format all source files
fmt:
    gofmt -w .

# Tidy go.mod and go.sum
tidy:
    go mod tidy

# Run the demo app
demo *args:
    go run ./cmd/demo {{args}}

# Run the showcase app
showcase *args:
    go run ./cmd/showcase {{args}}

# Run lazymake
lazymake *args:
    go run ./cmd/lazymake {{args}}

# Remove build artifacts
clean:
    rm -f coverage.out

# Release: merge develop into main, tag, push, return to develop, prepend new Unreleased section
# Usage: just release 2.0.0
release version:
    sed -i 's/## \[Unreleased\]/## v{{version}}/' CHANGELOG.md
    git add CHANGELOG.md
    git commit -m "chore: release v{{version}}"
    git checkout main
    git merge --no-ff develop -m "Release v{{version}}"
    git tag v{{version}}
    git push origin main
    git push origin v{{version}}
    git checkout develop
    sed -i 's/## v{{version}}/## [Unreleased]\n\n---\n\n## v{{version}}/' CHANGELOG.md
    git add CHANGELOG.md
    git commit -m "chore: start [Unreleased] section after v{{version}}"
