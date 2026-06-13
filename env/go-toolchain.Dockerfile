FROM golang:1.23-alpine

WORKDIR /src

RUN apk add --no-cache git
