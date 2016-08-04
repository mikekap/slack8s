FROM alpine

RUN apk --update upgrade && \
    apk add ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ADD slack8s slack8s

ENTRYPOINT ["/slack8s"]
