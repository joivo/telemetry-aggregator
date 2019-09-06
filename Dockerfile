FROM debian:stretch-slim

MAINTAINER emanueljoivo@lsd.ufcg.edu.br

WORKDIR /service

EXPOSE 8088

COPY telemetry-aggregator /service

RUN ["/bin/sh", "-c", "chmod +x telemetry-aggregator"]

ENTRYPOINT ["./telemetry-aggregator"]