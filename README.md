# wasp-ingest-rtmp

## Getting Started

### Requirements

This source code has a few development dependencies:

- golang 1.17
- docker (for testing purposes)
- make (for development scripts)
- [golangci-lint](https://golangci-lint.run/)

There is a makefile with common scripts for building, testing, and linting the source code.

### Building

```
make build
```

### Testing

```
make test
```
Docker
```
docker build -t nginx-rtmp .
docker run -p 1935:1935 -p 8080:8080 -it nginx-rtmp:latest
```
set rtmp streaming device to rtmp://<your ip>:1935/stream/whatever
```
ffplay -fflags nobuffer -analyzeduration 0 -fast rtmp://localhost:1935/stream/whatever
```
Stats are available on http://localhost:8080/stats

### Linting

```
make lint
```
