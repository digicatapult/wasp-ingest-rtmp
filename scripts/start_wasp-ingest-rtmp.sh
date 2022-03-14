#!/bin/sh

exec env $(cat /etc/envars | xargs) /bin/wasp-ingest-rtmp -rtmp rtmp://localhost:1935/$1/$2 >> /var/log/nginx/ingest.log