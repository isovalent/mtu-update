FROM alpine:3.7
ADD mtu-update /
ENTRYPOINT ["./mtu-update"]
