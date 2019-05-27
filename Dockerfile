FROM golang:1.12.5-alpine3.9 as builder

RUN apk update \
    && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

RUN adduser -D -g '' appuser

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ./dist/linux_386/kibana-config-controller /bin/kibana-config-controller

USER appuser
ENTRYPOINT ["/bin/kibana-config-controller"]
