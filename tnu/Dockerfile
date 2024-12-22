FROM golang:1.23-alpine AS builder
ARG VERSION
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG TARGETPLATFORM
ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}
RUN apk add --no-cache build-base
WORKDIR /app
COPY . .
RUN go build -a -ldflags "-s -w" -trimpath -o tnu .

FROM alpine:latest
RUN apk add --no-cache smartmontools
COPY --from=builder /app/tnu /usr/local/bin/tnu
ENTRYPOINT ["/usr/local/bin/tnu"]
