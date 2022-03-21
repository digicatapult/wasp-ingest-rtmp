FROM alpine AS production
RUN apk --no-cache add nginx-mod-rtmp ffmpeg
RUN mkdir /etc/nginx/stat/ /etc/nginx/rtmp.d/
COPY wasp-ingest-rtmp /bin/wasp-ingest-rtmp
COPY config/default.conf /etc/nginx/conf.d/default.conf
COPY config/nginx.conf /etc/nginx/nginx.conf
COPY config/stat.xsl /etc/nginx/stat/stat.xsl
COPY config/stream_prod.conf /etc/nginx/rtmp.d/stream_prod.conf
COPY scripts/start_wasp-ingest-rtmp.sh /bin/ingest-rtmp.sh
ENV ENV=production
EXPOSE 1935 8080
CMD ["sh", "-c", "/bin/printenv > /etc/envars ; /usr/sbin/nginx"]