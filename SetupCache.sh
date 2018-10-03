#!/bin/bash

# Environment variables
export user="web_cache_service"
export password="password"
export db="web_cache_db"


# install mongodb
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10

echo "deb http://repo.mongodb.org/apt/ubuntu xenial/mongodb-org/3.4 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.4.list

sudo apt-get update

sudo apt-get install -y mongodb-org

sudo vim /etc/systemd/system/mongodb.service << EOL
#Unit contains the dependencies to be satisfied before the service is started.
[Unit]
Description=MongoDB Database
After=network.target
Documentation=https://docs.mongodb.org/manual
# Service tells systemd, how the service should be started.
# Key `User` specifies that the server will run under the mongodb user and
# `ExecStart` defines the startup command for MongoDB server.
[Service]
User=mongodb
Group=mongodb
ExecStart=/usr/bin/mongod --quiet --config /etc/mongod.conf
# Install tells systemd when the service should be automatically started.
# `multi-user.target` means the server will be automatically started during boot.
[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload

sudo systemctl start mongodb

# create base user in mongodb
mongo --eval "use $db"
mongo --eval "db.createUser({user: "$user", pwd: "$password", roles:[{role:"dbOwner", db: "$db"}]})"

# get repo
sudo apt-get install -y git
cd ~
mkdir a2
cd a2
echo 'run git clone https://github.ugrad.cs.ubc.ca/CPSC416-2018W-T1/A2-c8z9a-q1v0b.git'

# get go

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


