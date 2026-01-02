FROM golang:1.25

WORKDIR /app

ENV TZ=Asia/Shanghai

RUN mkdir -p /app/configs
COPY dnsmasq_exporter /app/dnsmasq_exporter

ENTRYPOINT ["/app/dnsmasq_exporter"]