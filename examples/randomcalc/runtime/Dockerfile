FROM golang:alpine

ENV GODEBUG "netdns=cgo"
RUN mkdir -p "$GOPATH/src/github.com/lnsp/hive/examples/randomcalc/runtime"
ADD . "$GOPATH/src/github.com/lnsp/hive/examples/randomcalc/runtime"

WORKDIR "$GOPATH/src/github.com/lnsp/hive/examples/randomcalc/runtime"
RUN go get -v
RUN go build

CMD ["./runtime"]