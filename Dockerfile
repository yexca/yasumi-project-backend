FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/yasumi-api ./cmd/yasumi-api

FROM alpine:3.20

RUN adduser -D -H -u 10001 yasumi
USER yasumi

COPY --from=build /out/yasumi-api /usr/local/bin/yasumi-api

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/yasumi-api"]
