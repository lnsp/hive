FROM golang:alpine

ENV GODEBUG "netdns=cgo"

RUN mkdir -p "$GOPATH/src/github.com/lnsp/hive/examples/addition/runtime"
ADD . "$GOPATH/src/github.com/lnsp/hive/examples/addition/runtime"

WORKDIR "$GOPATH/src/github.com/lnsp/hive/examples/addition/runtime"
RUN go get -v
RUN go build

CMD ["./runtime"]