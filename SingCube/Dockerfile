# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:22-alpine AS web-builder
WORKDIR /build/web
COPY web/package*.json ./
RUN npm install --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS server-builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build/server
COPY server/go.mod ./
COPY server/ ./
COPY --from=web-builder /build/server/internal/webui/dist ./internal/webui/dist
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/singcube-server ./cmd/singcube

FROM --platform=$TARGETPLATFORM alpine:3.21 AS runtime
ARG TARGETARCH
RUN apk add --no-cache iptables ip6tables nftables iproute2 curl ca-certificates tzdata procps dropbear && \
    mkdir -p /etc/init.d /etc/dropbear /opt/SingCube/bin /opt/SingCube/container/bin /opt/SingCube/data /opt/SingCube/etc /opt/SingCube/defaults/etc && \
    rm -rf /var/cache/apk/*
WORKDIR /opt/SingCube
COPY --from=server-builder /out/singcube-server /opt/SingCube/bin/singcube-server
COPY container/ /opt/SingCube/container/
COPY entrypoint.sh /opt/SingCube/entrypoint.sh
COPY meta.json /opt/SingCube/meta.json
RUN cp /opt/SingCube/container/bin/singcube-ctl /etc/init.d/singcube && \
    ln -sf /opt/SingCube/container/bin/singcube-ctl /usr/bin/singcube && \
    chmod +x /opt/SingCube/bin/singcube-server /opt/SingCube/entrypoint.sh /opt/SingCube/container/bin/* /etc/init.d/singcube
EXPOSE 2026 53/udp 53/tcp 7853/udp 7853/tcp 7890
VOLUME ["/opt/SingCube/data", "/opt/SingCube/etc"]
ENTRYPOINT ["/opt/SingCube/entrypoint.sh"]
