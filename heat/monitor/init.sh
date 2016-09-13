#!/bin/sh
REPOSITORY_URL=https://github.com/obazavil/openstack-workload-transcoding.git
APP_DIR=github.com/obazavil/openstack-workload-transcoding

sudo apt-get update
sudo apt-get install git -y

# Installing GoLang environment
sudo curl -O https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Setting up GoLang workspace
mkdir $HOME/go-workspace
mkdir $HOME/go-workspace/bin
mkdir $HOME/go-workspace/pkg
mkdir $HOME/go-workspace/src
export GOPATH=$HOME/go-workspace

# Installing required libraries
cd $GOPATH/src
go get github.com/go-kit/kit/log
go get golang.org/x/net/context
go get github.com/go-resty/resty
go get github.com/gorilla/mux
go get github.com/rackspace/gophercloud

# Downloading the code application and running the monitor microservice
mkdir -p $APP_DIR
git clone $REPOSITORY_URL $APP_DIR
cd $APP_DIR
go run transcoding/monitor/cmd/main.go -database=https://$DATABASE_ENDPOINT:8080
