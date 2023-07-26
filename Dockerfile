FROM golang:1.20.6-alpine3.18 AS builder
WORKDIR /go/src/producer-rss
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM scratch
COPY --from=builder /go/src/producer-rss/producer-rss /bin/producer-rss
ENTRYPOINT ["/bin/producer-rss", "/etc/feed-urls.txt"]
