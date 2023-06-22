FROM golang:1.20.5 AS builder

COPY . $GOPATH/src/oauth2-proxy-nexus3/
WORKDIR "$GOPATH/src/oauth2-proxy-nexus3"
RUN CGO_ENABLED=0 go build -o /tmp/oauth2-proxy-nexus3

FROM scratch

COPY --from=builder /tmp/oauth2-proxy-nexus3 /

ENTRYPOINT [ "/oauth2-proxy-nexus3" ]
