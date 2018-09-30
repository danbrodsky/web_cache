#!/bin/bash

root_file=/run/secrets/mongo-root-user
password_file=/run/secrets/mongo-root-password
user_password_file=/run/secrets/mongo-user-password

root="$(< "${root_file}")"
password="$(< "${password_file}")"
user_password="$(< "${user_password_file}")"

services=("web_cache")

for service in ${services[@]}
do
    db="${service}_db"
    user="${service}_service"
    mongo -u $root -p $password --authenticationDatabase admin <<EOF
use $db
db.createUser({user: '$user', pwd: '$user_password', roles:[{role:'dbOwner', db:'$db'}]})
EOF
done

