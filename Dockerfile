FROM golang:1.17-alpine AS build

WORKDIR /wasp-ingest-rtmp
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wasp-ingest-mqtt

FROM alpine
COPY --from=build /bin/wasp-ingest-mqtt /bin/wasp-ingest-mqtt
ENTRYPOINT ["/bin/wasp-ingest-mqtt"]