FROM golang:1.23 as builder
WORKDIR /build
COPY ./ /build/
RUN apt-get update \
    && apt-get install make \
    && make clean \
    && make test \
    && make build

FROM scratch

COPY --from=builder /build/bin/* /
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT [ "/tgbot" ]