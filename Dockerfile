FROM alpine:3.9

RUN apk update \
    && apk add --no-cache curl \
                          ca-certificates \
                          tzdata \
    && update-ca-certificates

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator
USER kube-operator

COPY kibana-config-controller /bin/kibana-config-controller
ENTRYPOINT ["/bin/kibana-config-controller"]
