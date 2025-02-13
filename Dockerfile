FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
COPY ./ /build/
RUN apt-get update \
    && apt-get --no-install-recommends -y install make=4.3-4.1
RUN make clean \
    && make test \
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

RUN echo "nobody:x:65534:65534:Nobody:/:" > /build/passwd

FROM scratch

COPY --from=builder /build/bin/* /
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /build/passwd /etc/passwd
COPY config/miaou.yaml /config/miaou.yaml

USER nobody
ENTRYPOINT [ "/miaou" ]
