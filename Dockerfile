# syntax=docker/dockerfile:1.7

FROM golang:1.25.12-alpine3.24 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG SERVICE=app-apis

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && \
    case "${SERVICE}" in \
      app-apis) package=./app/apis/cmd/app-apis ;; \
      user-rpc) package=./app/user/rpc ;; \
      interview-rpc) package=./app/interview/rpc ;; \
      question-rpc) package=./app/question/rpc ;; \
      report-worker) package=./app/pump/cmd/report-worker ;; \
      *) echo "unsupported SERVICE: ${SERVICE}" >&2; exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service "${package}"

FROM alpine:3.24

ARG SERVICE=app-apis

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S whetstone && \
    adduser -S -G whetstone whetstone && \
    mkdir -p /etc/whetstone /srv/whetstone && \
    chown -R whetstone:whetstone /srv/whetstone

COPY --from=builder --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
COPY deploy/dokploy/etc/app-apis.yaml /etc/whetstone/app-apis.yaml
COPY deploy/dokploy/etc/user.yaml /etc/whetstone/user.yaml
COPY deploy/dokploy/etc/interview.yaml /etc/whetstone/interview.yaml
COPY deploy/dokploy/etc/question.yaml /etc/whetstone/question.yaml
COPY --chmod=755 deploy/dokploy/entrypoint.sh /usr/local/bin/whetstone-entrypoint

ENV TZ=Asia/Shanghai \
    CONFIG_SOURCE=env \
    WHETSTONE_SERVICE=${SERVICE}

USER whetstone
WORKDIR /srv/whetstone

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/local/bin/whetstone-entrypoint"]
