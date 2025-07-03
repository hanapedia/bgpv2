FROM ubuntu:jammy

# Install BIRD and common network tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        # bird2 \
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


ENTRYPOINT ["rivers", "--upstreams=10.0.2.1:6443", "--listen=127.0.0.1:16443", "--dial-timeout=4s", "--dial-keep-alive=6s", "--check-interval=5s"]
