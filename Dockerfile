FROM golang:1.12.5-alpine3.9 as builder

# Pass in proxy for pipeline
ARG HTTP_PROXY_ARG
ENV http_proxy=$HTTP_PROXY_ARG
ENV https_proxy=$HTTP_PROXY_ARG

RUN apk update \
    && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

RUN adduser -D -g '' appuser

RUN mkdir /build
WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/kibana-config-controller ./cmd

FROM scratch
USER kube-operator
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/kibana-config-controller /bin/kibana-config-controller

USER appuser
ENTRYPOINT ["/bin/kibana-config-controller"]
