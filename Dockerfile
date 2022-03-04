FROM golang:1.17-alpine AS build

WORKDIR /wasp-ingest-rtmp
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wasp-ingest-rtmp

FROM alpine
COPY --from=build /bin/wasp-ingest-rtmp /bin/wasp-ingest-rtmp
ENTRYPOINT ["/bin/wasp-ingest-rtmp"]