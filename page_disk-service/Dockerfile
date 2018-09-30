From ubuntu:16.04

ADD ./azureinstall.sh /

RUN  apt-get update \
  && apt-get install -y wget \
  && rm -rf /var/lib/apt/lists/* \
  && apt-get update \
  && apt-get -y install sudo 

RUN ./azureinstall.sh

RUN apt-get install -y vim
RUN apt-get install -y git

ADD /go /root/go

EXPOSE 8888
EXPOSE 8000

CMD tail -f /dev/null
