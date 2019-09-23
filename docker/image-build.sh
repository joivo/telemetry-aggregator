#!/bin/sh

if [[ "$#" -ne 1 ]]; then
  echo "Usage: $0 <docker tag>"
  exit 1
fi

tag=$1

docker build -t emanueljoivo/telemetry-aggregator:${tag} .