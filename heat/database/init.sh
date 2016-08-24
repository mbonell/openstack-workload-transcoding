#!/bin/sh
REPOSITORY_URL=https://github.com/obazavil/openstack-workload-transcoding.git
APP_DIR=github.com/obazavil/openstack-workload-transcoding

# Preparing the empty block device
sudo mke2fs /dev/vdb

# Mount the block device to the MongoDB database files
mkdir /var/lib/mongodb
sudo echo "/dev/vdb /var/lib/mongodb ext4 defaults  1 2" >> /etc/fstab
sudo mount /var/lib/mongodb

sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927
echo "deb http://repo.mongodb.org/apt/ubuntu trusty/mongodb-org/3.2 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.2.list
sudo apt-get update
sudo apt-get install mongodb-org -y
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
go get gopkg.in/mgo.v2

# Downloading the code application and running the database microservice
mkdir -p $APP_DIR
git clone $REPOSITORY_URL $APP_DIR
cd $APP_DIR
go run database/cmd/main.go