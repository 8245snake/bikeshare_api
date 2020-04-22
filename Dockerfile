FROM golang:latest

WORKDIR /go/src/github.com

RUN apt-get update && apt-get install -y \
	git \
	--no-install-recommends

RUN go get github.com/kr/godep

ADD ./start.sh /go/src/
RUN chmod -R 777 /go/src/start.sh

CMD ["/bin/sh", "/go/src/start.sh"]