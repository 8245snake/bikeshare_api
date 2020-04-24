FROM golang:latest

WORKDIR /usr/

RUN apt-get update && apt-get install -y \
    bash \
    git \
    --no-install-recommends
RUN go get github.com/kr/godep

ADD ./start.sh /go/
RUN chmod 777 /go/start.sh

CMD ["bash", "/go/start.sh"]