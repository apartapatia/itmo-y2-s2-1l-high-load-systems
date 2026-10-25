FROM golang:1.26.0 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -o /out/server .

FROM alpine:3.24.3 AS runtime

COPY --from=build /out/server /server
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]
