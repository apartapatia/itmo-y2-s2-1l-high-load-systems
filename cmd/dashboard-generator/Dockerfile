FROM golang:1.26.0 AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/dashboard-generator/*.go ./cmd/dashboard-generator/
RUN CGO_ENABLED=0 go build -trimpath -o /out/dashboard-generator ./cmd/dashboard-generator

FROM alpine:3.24.3
COPY --from=build /out/dashboard-generator /dashboard-generator
USER 65532:65532
ENTRYPOINT ["/dashboard-generator"]
