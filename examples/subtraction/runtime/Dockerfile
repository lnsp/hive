FROM golang:alpine

ENV GODEBUG "netdns=cgo"

RUN mkdir -p "$GOPATH/src/github.com/lnsp/hive/examples/subtraction/runtime"
ADD . "$GOPATH/src/github.com/lnsp/hive/examples/subtraction/runtime"

WORKDIR "$GOPATH/src/github.com/lnsp/hive/examples/subtraction/runtime"
RUN go get -v
RUN go build

CMD ["./runtime"]