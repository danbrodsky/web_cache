#!/bin/bash

#go to home directory
cd

#download go binary
wget https://storage.googleapis.com/golang/go1.9.7.linux-amd64.tar.gz

#unzip and remove
sudo tar -C /usr/local -xzf go1.9.7.linux-amd64.tar.gz
rm go1.9.7.linux-amd64.tar.gz

sed -i -re 's/^(mesg n)(.*)$/#\1\2/g' /root/.profile \
source /root/.profile

#install mercurial
sudo apt-get install mercurial -y

# install GoVector to boot strap directories
# go get github.com/arcaneiceman/GoVector

source .profile

