# syntax=docker/dockerfile:1.7

FROM mcr.microsoft.com/dotnet/sdk:10.0.300 AS build

ARG TARGETARCH

WORKDIR /src

COPY src/gateway/global.json src/gateway/Directory.Build.props src/gateway/Archivist.Gateway.slnx ./src/gateway/
COPY src/gateway/Archivist.Gateway.Api/Archivist.Gateway.Api.csproj ./src/gateway/Archivist.Gateway.Api/
COPY src/gateway/Archivist.Gateway.Application/Archivist.Gateway.Application.csproj ./src/gateway/Archivist.Gateway.Application/

RUN --mount=type=cache,target=/root/.nuget/packages \
    case "${TARGETARCH}" in \
        amd64) dotnet_runtime=linux-x64 ;; \
        arm64) dotnet_runtime=linux-arm64 ;; \
        *) echo "Unsupported target architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && dotnet restore src/gateway/Archivist.Gateway.Api/Archivist.Gateway.Api.csproj --runtime "${dotnet_runtime}"

COPY src/gateway/ ./src/gateway/

RUN --mount=type=cache,target=/root/.nuget/packages \
    case "${TARGETARCH}" in \
        amd64) dotnet_runtime=linux-x64 ;; \
        arm64) dotnet_runtime=linux-arm64 ;; \
        *) echo "Unsupported target architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && dotnet publish src/gateway/Archivist.Gateway.Api/Archivist.Gateway.Api.csproj \
        --configuration Release \
        --runtime "${dotnet_runtime}" \
        --self-contained true \
        --no-restore \
        -p:PublishSingleFile=true \
        -p:PublishAot=false \
        -p:IncludeNativeLibrariesForSelfExtract=false \
        --output /app/publish

RUN mkdir -p /runtime/data \
    && find /app/publish -type d -exec chmod 0555 {} + \
    && find /app/publish -type f -exec chmod 0444 {} + \
    && chmod 0555 /app/publish/Archivist.Gateway.Api \
    && chmod 0755 /runtime/data

FROM mcr.microsoft.com/dotnet/runtime-deps:10.0-noble-chiseled AS runtime

ENV ASPNETCORE_URLS=http://0.0.0.0:8080
ENV DOTNET_URLS=http://0.0.0.0:8080
ENV ASPNETCORE_ENVIRONMENT=Production
ENV DOTNET_ENVIRONMENT=Production
ENV ARCHIVIST_SQLITE_PATH=/data/archive.db
ENV ARCHIVIST_DATA_DIR=/data

WORKDIR /app

COPY --from=build --chown=0:0 /app/publish/ /app/
COPY --from=build --chown=65532:65532 /runtime/data/ /data/

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["./Archivist.Gateway.Api"]
