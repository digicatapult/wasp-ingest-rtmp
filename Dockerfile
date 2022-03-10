FROM golang:1.17-alpine AS build

WORKDIR /wasp-ingest-rtmp
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/wasp-ingest-rtmp

FROM alpine
RUN apk --no-cache add nginx-mod-rtmp
RUN mkdir /etc/nginx/stat/
COPY --from=build /bin/wasp-ingest-rtmp /bin/wasp-ingest-rtmp
COPY ./config/default.conf /etc/nginx/conf.d/default.conf
COPY ./config/nginx.conf /etc/nginx/nginx.conf
COPY ./config/stat.xsl /etc/nginx/stat/stat.xsl
EXPOSE 1935 8080
CMD ["/usr/sbin/nginx"]