FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/yasumi-api ./cmd/yasumi-api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/yasumi-migrate ./cmd/yasumi-migrate

FROM alpine:3.20

ARG YASUMI_IMAGE_REPOSITORY=yexca/yasumi-backend
ARG YASUMI_IMAGE_TAG=0.1.1

LABEL org.opencontainers.image.title="yasumi-backend" \
      org.opencontainers.image.vendor="yexca" \
      org.opencontainers.image.version="${YASUMI_IMAGE_TAG}" \
      org.opencontainers.image.ref.name="${YASUMI_IMAGE_REPOSITORY}:${YASUMI_IMAGE_TAG}"

RUN apk add --no-cache tzdata

RUN adduser -D -H -u 10001 yasumi
USER yasumi

COPY --from=build /out/yasumi-api /usr/local/bin/yasumi-api
COPY --from=build /out/yasumi-migrate /usr/local/bin/yasumi-migrate

EXPOSE 7659
ENTRYPOINT ["/usr/local/bin/yasumi-api"]
