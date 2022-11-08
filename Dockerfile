FROM golang:1.16 AS build
WORKDIR /sloop
COPY go.mod go.sum ./
RUN go mod download
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -installsuffix cgo -o sloop ./pkg/sloop
FROM gcr.io/distroless/base
COPY --from=build /sloop/sloop /sloop
COPY --from=build /sloop/pkg/sloop/webserver/webfiles /pkg/sloop/webserver/webfiles
CMD ["/sloop"]