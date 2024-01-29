FROM --platform=$BUILDPLATFORM golang:1.21.6-alpine3.18@sha256:3354c3a94c3cf67cb37eb93a8e9474220b61a196b13c26f1c01715c301b22a69 AS builder

RUN apk add --update --no-cache ca-certificates make git curl

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

WORKDIR /usr/local/src/portward

ARG GOPROXY

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /usr/local/bin/portward .

FROM alpine:3.19.0@sha256:51b67269f354137895d43f3b3d810bfacd3945438e94dc5ac55fdac340352f48 AS alpine

RUN apk add --update --no-cache ca-certificates tzdata bash

SHELL ["/bin/bash", "-c"]

COPY --from=builder /usr/local/bin/portward /usr/local/bin/

CMD portward
