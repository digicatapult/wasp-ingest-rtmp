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

## Usage

```
$ ./wasp-ingest-rtmp --help
Usage of ./wasp-ingest-rtmp:
  -rtmp string
    	The url of the rtmp stream to ingest (default "default")
```

### Log Levels

Logging levels can be configured when the application is running in "production" mode. To do this a couple of environment variables need to be set.

```
$ ENV=production LOG_LEVEL=warn ./wasp-ingest-rtmp
```
