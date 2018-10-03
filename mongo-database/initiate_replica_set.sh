#!/bin/bash

# Name of the mongodb container
container=$1

root_file=/run/secrets/mongo-root-user
password_file=/run/secrets/mongo-root-password

root=`cat "$root_file"`
password=`cat "$password_file"`

mongo -u $root -p $password --authenticationDatabase admin <<EOF
rs.initiate()
exit
EOF
