FROM golang:1.7.4-wheezy

RUN go get github.com/Beh01der/go-log-monit-service && rm -rf $GOPATH/pkg/* && rm -rf $GOPATH/src/*

ENTRYPOINT ["$GOPATH/bin/go-log-monit-service"]
