# Build Instructions

## Prerequisites

- Go 1.22 or later
- (Optional) `golangci-lint` for linting

## Build

```sh
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
go build -ldflags "-X main.Version=$VERSION" -o basicc ./cmd/basicc
```

## Install

```sh
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
go install -ldflags "-X main.Version=$VERSION" ./cmd/basicc
```

## Test

```sh
go test -v -cover ./...
```

## Code Coverage

```sh
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Lint

```sh
golangci-lint run ./...
```

## Vet

```sh
go vet ./...
```

## Format

```sh
gofmt -s -w .
```

## Clean

```sh
rm -f basicc coverage.out
```
