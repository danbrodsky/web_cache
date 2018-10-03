docker-compose -p deploy build
docker-compose -p deploy up -d --no-recreate mongodb
sleep 1
docker-compose -p deploy up -d --no-recreate webcache-service

docker rmi $(docker images | grep "^<none>" | awk '{print $3}') &> /dev/null
docker volume rm $(docker volume ls -qf dangling=true) &> /dev/null
