FROM golang:1.7.4-wheezy

RUN go get github.com/Beh01der/go-log-monit-service && rm -rf /go/pkg/* && rm -rf /go/src/*

ENTRYPOINT ["/go/bin/go-log-monit-service"]
