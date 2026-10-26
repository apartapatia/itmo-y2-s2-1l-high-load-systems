FROM golang:1.26.0 AS build
WORKDIR /src
RUN git clone --depth 1 --branch v0.6.2 https://github.com/yandex/pandora.git . \
    && CGO_ENABLED=0 go build -trimpath -o /out/pandora .

FROM alpine:3.24.3
COPY --from=build /out/pandora /usr/local/bin/pandora
WORKDIR /load
COPY load.yaml scenarios.yaml run.sh ./
USER 65532:65532
ENTRYPOINT ["/bin/sh", "/load/run.sh"]
