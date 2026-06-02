# syntax=docker/dockerfile:1.7

FROM node:22-bookworm AS ui-build

WORKDIR /src/ui

COPY src/ui/package.json src/ui/package-lock.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY src/ui/ ./

ARG VERSION_LABEL=abc123
ENV VITE_VERSION_LABEL=${VERSION_LABEL}

ARG VITE_API_BASE_PATH=/api
ENV VITE_API_BASE_PATH=${VITE_API_BASE_PATH}

RUN --mount=type=cache,target=/root/.npm \
    npm run build

RUN cp -a dist /out \
    && find /out -type d -exec chmod 0555 {} + \
    && find /out -type f -exec chmod 0444 {} +

FROM --platform=$BUILDPLATFORM caddy:2.11.2-builder AS caddy-build

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    case "${TARGETARCH}" in \
        amd64) go_arch=amd64 ;; \
        arm64) go_arch=arm64 ;; \
        *) echo "Unsupported target architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && CGO_ENABLED=0 GOOS=linux GOARCH="${go_arch}" \
        xcaddy build v2.11.2 --output /out/caddy

RUN mkdir -p /out/etc/caddy /out/var/lib/caddy \
    && chmod 0555 /out/caddy \
    && chmod 0555 /out/etc/caddy \
    && chmod 0755 /out/var/lib/caddy

COPY ui.Caddyfile /out/etc/caddy/Caddyfile

RUN chmod 0444 /out/etc/caddy/Caddyfile

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

ENV HOME=/var/lib/caddy
ENV XDG_DATA_HOME=/var/lib/caddy

WORKDIR /usr/share/archivist/ui

COPY --from=caddy-build --chown=0:0 /out/caddy /usr/bin/caddy
COPY --from=caddy-build --chown=0:0 /out/etc/caddy/ /etc/caddy/
COPY --from=caddy-build --chown=65532:65532 /out/var/lib/caddy/ /var/lib/caddy/
COPY --from=ui-build --chown=0:0 /out/ /usr/share/archivist/ui/

USER nonroot

EXPOSE 8080

VOLUME ["/var/lib/caddy"]

ENTRYPOINT ["/usr/bin/caddy", "run", "--config", "/etc/caddy/Caddyfile"]
