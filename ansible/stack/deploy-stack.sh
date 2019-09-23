#!/bin/sh

sudo docker stack deploy -c $(pwd)/stack/docker-stack.yml monitoring
