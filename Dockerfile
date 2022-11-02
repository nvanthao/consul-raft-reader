FROM golang:1.18-alpine AS builder
COPY . /build
WORKDIR /build
RUN go get .
RUN go build -v -o /tmp/consul-raft-reader

FROM alpine:3.16.2 AS final
COPY --from=builder /tmp/consul-raft-reader /usr/local/bin/consul-raft-reader
ENTRYPOINT ["/usr/local/bin/consul-raft-reader"]