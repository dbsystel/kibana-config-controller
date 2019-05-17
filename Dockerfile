FROM alpine:latest

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator

USER kube-operator

COPY ./bin/kibana-config-controller .

ENTRYPOINT ["./kibana-config-controller"]
