# syntax=docker/dockerfile:1.7

FROM python:3.12-slim-bookworm AS build

ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy
ENV UV_PROJECT_ENVIRONMENT=/app/venv

WORKDIR /src/snapshotter

COPY --from=ghcr.io/astral-sh/uv:0.9.17 /uv /uvx /usr/local/bin/

COPY src/snapshotter/pyproject.toml src/snapshotter/uv.lock ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-install-project

COPY src/snapshotter/ ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-editable \
    && mkdir -p /runtime/data /runtime/tmp/archivist-snapshotter /runtime/work/archivist-snapshotter \
    && find /app/venv -type d -exec chmod 0555 {} + \
    && find /app/venv -type f -exec chmod 0444 {} + \
    && find /app/venv/bin -type f -exec chmod 0555 {} + \
    && chmod 0755 /runtime/data /runtime/tmp /runtime/tmp/archivist-snapshotter /runtime/work /runtime/work/archivist-snapshotter

FROM gcr.io/distroless/base-debian12:nonroot AS runtime

ENV PATH=/app/venv/bin:/usr/local/bin
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=
ENV ARCHIVIST_DATA_DIR=/data
ENV ARCHIVIST_SQLITE_PATH=/data/archive.db
ENV ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS=86400
ENV ARCHIVIST_SNAPSHOTTER_WORK_DIR=/tmp/archivist-snapshotter
ENV ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX=

WORKDIR /app

COPY --from=build --chown=0:0 /usr/local/ /usr/local/
COPY --from=build --chown=0:0 /lib/ /lib/
COPY --from=build --chown=0:0 /usr/lib/ /usr/lib/
COPY --from=build --chown=0:0 /app/venv/ /app/venv/
COPY --from=build --chown=65532:65532 /runtime/data/ /data/
COPY --from=build --chown=65532:65532 /runtime/tmp/ /tmp/
COPY --from=build --chown=65532:65532 /runtime/work/ /work/

USER nonroot

VOLUME ["/data", "/work"]

ENTRYPOINT ["/app/venv/bin/archivist-snapshotter"]
