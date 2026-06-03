from __future__ import annotations

from pathlib import Path
from typing import Protocol, cast

import boto3

from archivist_snapshotter.config import S3Config
from archivist_snapshotter.logging import JsonLogger
from archivist_snapshotter.telemetry import get_tracer, mark_span_error


class S3Client(Protocol):
    def upload_file(self, filename: str, bucket: str, key: str) -> object: ...


class S3Factory(Protocol):
    def __call__(self, service_name: str, **kwargs: object) -> S3Client: ...


class S3Uploader:
    def __init__(
        self,
        config: S3Config,
        logger: JsonLogger,
        *,
        client_factory: S3Factory | None = None,
    ) -> None:
        self._config = config
        self._logger = logger
        self._client_factory = (
            cast(S3Factory, boto3.client) if client_factory is None else client_factory
        )

    def upload(self, archive_path: Path, object_key: str) -> None:
        with get_tracer().start_as_current_span(
            "snapshotter.s3_upload",
            attributes={
                "archive.path": str(archive_path),
                "s3.bucket": self._config.bucket,
                "s3.object_key": object_key,
                "s3.endpoint_url": self._config.endpoint_url,
                "s3.region": self._config.region,
            },
            record_exception=False,
            set_status_on_exception=False,
        ) as span:
            try:
                client = self._client_factory(
                    "s3",
                    endpoint_url=self._config.endpoint_url,
                    region_name=self._config.region,
                    aws_access_key_id=self._config.access_key_id,
                    aws_secret_access_key=self._config.secret_access_key,
                )
                client.upload_file(str(archive_path), self._config.bucket, object_key)
            except Exception:
                mark_span_error(span)
                raise
        self._logger.info(
            "upload_succeeded",
            archive_path=str(archive_path),
            bucket=self._config.bucket,
            object_key=object_key,
            endpoint_url=self._config.endpoint_url,
            region=self._config.region,
        )
