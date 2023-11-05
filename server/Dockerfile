#Certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# Build
FROM golang as base

WORKDIR /src

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /app

### App
FROM scratch as app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /app /
ENTRYPOINT ["/app"]
