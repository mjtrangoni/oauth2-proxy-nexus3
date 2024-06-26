FROM golang:1.22 AS builder

COPY . $GOPATH/src/oauth2-proxy-nexus3/
WORKDIR "$GOPATH/src/oauth2-proxy-nexus3"
RUN CGO_ENABLED=0 go build -o /tmp/oauth2-proxy-nexus3

FROM scratch
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/mjtrangoni/oauth2-proxy-nexus3"

COPY --from=builder /tmp/oauth2-proxy-nexus3 /

ENTRYPOINT [ "/oauth2-proxy-nexus3" ]
