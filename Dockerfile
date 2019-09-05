FROM alpine

WORKDIR /service

EXPOSE 8088

COPY aggregator /service

ENTRYPOINT ["sh", "-c", "'./aggregator'"]