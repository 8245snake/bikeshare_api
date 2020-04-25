FROM golang:latest

WORKDIR /usr/
ADD ./docker-entrypoint.sh .
RUN chmod 777 ./docker-entrypoint.sh

RUN apt-get update && apt-get install -y \
    bash \
    git \
    --no-install-recommends
RUN go get github.com/kr/godep

RUN cd /go/src/github.com && \
    git clone https://github.com/8245snake/bikeshare_api.git && \
    chmod -R 777 /go/src/github.com

CMD ["bash", "/usr/docker-entrypoint.sh"]