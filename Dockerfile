FROM golang:1.6.2

RUN mkdir -p $GOPATH/src/github.com/sfavron/slack8s

WORKDIR $GOPATH/src/github.com/sfavron/slack8s
ADD . $GOPATH/src/github.com/sfavron/slack8s

RUN go install

CMD ["slack8s"]
