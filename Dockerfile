FROM golang:1.23 AS builder
WORKDIR /build
COPY ./ /build/
RUN apt-get update \
    && apt-get --no-install-recommends -y install make=4.3-4.1 \
    && make clean \
    && make test \
    && make build

FROM scratch

COPY --from=builder /build/bin/* /
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT [ "/miaou" ]
