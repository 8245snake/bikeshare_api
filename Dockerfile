FROM golang:latest

WORKDIR /usr/

RUN apt-get update && apt-get install -y \
	git \
	--no-install-recommends
RUN go get github.com/kr/godep

ADD ./start.sh /go/
RUN chmod 777 /go/start.sh

CMD ["/bin/sh", "/go/start.sh"]