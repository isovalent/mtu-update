FROM golang:1.10.3 as vendor
WORKDIR /go/src/github.com/cilium/mtu-update/
COPY vendor/ .

FROM vendor as builder
WORKDIR /go/src/github.com/cilium/mtu-update/
COPY ./ .
# Exclude vendor/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o mtu-update .

FROM alpine:3.7
COPY --from=builder /go/src/github.com/cilium/mtu-update/mtu-update /
ENTRYPOINT ["./mtu-update"]
