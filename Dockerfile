FROM golang:1.12.5-alpine3.9 as builder

RUN apk update \
    && apk add --no-cache git \
                          ca-certificates \
                          tzdata \
                          curl \
    && update-ca-certificates

RUN adduser -D -g '' appuser

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/bin/curl /usr/bin/curl
COPY kibana-config-controller /bin/kibana-config-controller

USER appuser
ENTRYPOINT ["/bin/kibana-config-controller"]
