FROM golang:alpine

WORKDIR /service

EXPOSE 8090

RUN apk add git

RUN apk add curl

ENV SRC_DIR=/go/src/github.com/emanueljoivo/telemetry-aggregator

ENV CGO_ENABLED 0

ADD . $SRC_DIR

RUN cd $SRC_DIR; go mod tidy; go test

RUN cd $SRC_DIR; go build -o aggregator; cp aggregator /service/

ENTRYPOINT ["./aggregator"]