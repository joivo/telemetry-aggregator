FROM golang:rc-alpine

WORKDIR /service

EXPOSE 8090

RUN apk add git

RUN apk add curl

ENV SRC_DIR=/go/src/github.com/emanueljoivo/telemetry-aggregator

ADD . $SRC_DIR

RUN cd $SRC_DIR; chmod +x build.sh && ./build.sh

RUN cd $SRC_DIR; go build -o aggregator; cp aggregator /service/

ENTRYPOINT ["./aggregator"]