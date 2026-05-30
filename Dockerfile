# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" -o /out/genupdate-server ./cmd/main

FROM alpine:3.22

RUN addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/update /app/log \
    && touch /app/update/.ignore \
    && chown -R app:app /app

WORKDIR /app

COPY --from=build /out/genupdate-server /usr/local/bin/genupdate-server

USER app

ENV GIN_MODE=release \
    GENUPDATE_PORT=8090 \
    GENUPDATE_UPDATE_DIR=/app/update \
    GENUPDATE_MAX_CONCURRENT_DOWNLOADS=64

EXPOSE 8090

VOLUME ["/app/update", "/app/log"]

ENTRYPOINT ["genupdate-server"]
