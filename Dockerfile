# syntax=docker/dockerfile:1.7

FROM golang:1.25.12-alpine3.24 AS builder-base

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

FROM builder-base AS builder-app-apis
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service ./app/apis/cmd/app-apis

FROM builder-base AS builder-user-rpc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service ./app/user/rpc

FROM builder-base AS builder-interview-rpc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service ./app/interview/rpc

FROM builder-base AS builder-question-rpc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service ./app/question/rpc

FROM builder-base AS builder-report-worker
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/whetstone-service ./app/pump/cmd/report-worker

FROM alpine:3.24 AS runtime-base

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S whetstone && \
    adduser -S -G whetstone whetstone && \
    mkdir -p /etc/whetstone /srv/whetstone && \
    chown -R whetstone:whetstone /srv/whetstone

COPY deploy/dokploy/etc/app-apis.yaml /etc/whetstone/app-apis.yaml
COPY deploy/dokploy/etc/user.yaml /etc/whetstone/user.yaml
COPY deploy/dokploy/etc/interview.yaml /etc/whetstone/interview.yaml
COPY deploy/dokploy/etc/question.yaml /etc/whetstone/question.yaml
COPY --chmod=755 deploy/dokploy/entrypoint.sh /usr/local/bin/whetstone-entrypoint

ENV TZ=Asia/Shanghai \
    CONFIG_SOURCE=env

USER whetstone
WORKDIR /srv/whetstone

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/local/bin/whetstone-entrypoint"]

FROM runtime-base AS app-apis
COPY --from=builder-app-apis --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
ENV WHETSTONE_SERVICE=app-apis
EXPOSE 8888
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8888/healthz || exit 1

FROM runtime-base AS user-rpc
COPY --from=builder-user-rpc --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
ENV WHETSTONE_SERVICE=user-rpc
EXPOSE 9001

FROM runtime-base AS interview-rpc
COPY --from=builder-interview-rpc --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
ENV WHETSTONE_SERVICE=interview-rpc
EXPOSE 9002

FROM runtime-base AS question-rpc
COPY --from=builder-question-rpc --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
ENV WHETSTONE_SERVICE=question-rpc
EXPOSE 9003

FROM runtime-base AS report-worker
COPY --from=builder-report-worker --chown=whetstone:whetstone /out/whetstone-service /usr/local/bin/whetstone-service
ENV WHETSTONE_SERVICE=report-worker
