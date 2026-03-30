#!/bin/sh
set -e

SCHEDULE="${NEBULA_DNS_SCHEDULE:-*/5 * * * *}"

echo "$SCHEDULE /usr/local/bin/nebula-dns" > /etc/crontabs/root

/usr/local/bin/nebula-dns
echo "Starting crond with schedule: $SCHEDULE"
exec crond -f -l 2
