# Development

## Prerequisites

* [Go](https://go.dev/) 1.24
* [GoReleaser](https://goreleaser.com/) 2.7.0+

## How to build

```bash
go test ./...
go build
```

## How to create release binary at local

```bash
goreleaser release --snapshot --clean
```

## How to upgrade dependent modules

```bash
go get -u
go mod tidy
```
