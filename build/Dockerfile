FROM golang:1.25

WORKDIR /app

ENV TZ=Asia/Shanghai

RUN mkdir -p /app/configs
COPY dnsmasq_gateway /app/dnsmasq_gateway

ENTRYPOINT ["/app/dnsmasq_gateway"]