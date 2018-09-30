FROM mongo:3.6.4

COPY ./mongo-root-user /run/secrets/mongo-root-user
COPY ./mongo-root-password /run/secrets/mongo-root-password
COPY ./mongo-user-password /run/secrets/mongo-user-password
COPY ./setup.sh /docker-entrypoint-initdb.d
COPY ./initiate_replica_set.sh /initiate_replica_set.sh

ENV MONGO_INITDB_ROOT_PASSWORD_FILE /run/secrets/mongo-root-password
ENV MONGO_INITDB_ROOT_USERNAME_FILE /run/secrets/mongo-root-user
