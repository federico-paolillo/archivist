# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS build

ARG TARGETARCH

WORKDIR /src/worker

COPY src/worker/go.mod src/worker/go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY src/worker/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    case "${TARGETARCH}" in \
        amd64) go_arch=amd64 ;; \
        arm64) go_arch=arm64 ;; \
        *) echo "Unsupported target architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && CGO_ENABLED=0 GOOS=linux GOARCH="${go_arch}" \
        go build -trimpath -ldflags="-s -w" -o /out/archivist-worker ./cmd/app

RUN mkdir -p /runtime/data \
    && chmod 0555 /out/archivist-worker \
    && chmod 0755 /runtime/data

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

ENV ARCHIVIST_SQLITE_PATH=/data/archive.db
ENV ARCHIVIST_DATA_DIR=/data
ENV ARCHIVIST_LLM_PROVIDER=anthropic
ENV ARCHIVIST_LLM_MODEL=claude-haiku-4-5-20251001

WORKDIR /app

COPY --from=build --chown=0:0 /out/archivist-worker /usr/bin/archivist-worker
COPY --from=build --chown=65532:65532 /runtime/data/ /data/

USER nonroot

VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/archivist-worker"]
CMD ["process"]
