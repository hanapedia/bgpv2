FROM ubuntu:jammy

# Install BIRD and common network tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        bird2 \
        iproute2 \
        tcpdump \
        inetutils-ping \
        inetutils-traceroute \
        dnsutils \
        net-tools \
        busybox \
        iptables \
        curl \
        ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Ensure /run/bird exists at runtime
RUN mkdir -p /run/bird

ENTRYPOINT ["/usr/sbin/bird", "-c", "/etc/bird.conf", "-d"]
