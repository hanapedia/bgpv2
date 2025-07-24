#!/bin/bash
set -e

# Run your pre-start binary
/tools/rivers --upstreams=10.0.1.2:6443 --listen=127.0.0.1:16443 --dial-timeout=4s --dial-keep-alive=6s --check-interval=5s &

# Now exec to bird to forward signals properly
exec /usr/sbin/bird -c /etc/bird.conf -d
