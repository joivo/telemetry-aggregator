#!/bin/bash

# A script to setup a swarm manager environment.
# It expects to be run on Ubuntu 16.04 via 'sudo'

install-docker() {
    echo "--> Installing docker"
    apt update

    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        software-properties-common
    
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

    apt-key fingerprint 0EBFCD88

    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) \
    stable"

    apt-get update

    apt-get install -y docker-ce
}

run() {
    CHECK_DOCKER_INSTALLATION=$(dpkg -l | grep -c docker-ce)

    if ! [ $CHECK_DOCKER_INSTALLATION -ne 0 ]; then
        install-docker
    else 
        echo "--> Docker its already installed"
    fi    

    docker swarm init    
}

run