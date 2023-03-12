FROM golang:1.20.2-alpine3.17 AS builder
WORKDIR /go/src/producer-rss
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.17.0
COPY --from=builder /go/src/producer-rss/producer-rss /bin/producer-rss
ENTRYPOINT ["/bin/producer-rss", "/etc/feeds.txt"]
