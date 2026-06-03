from __future__ import annotations

import io
from pathlib import Path

from archivist_snapshotter.logging import JsonLogger
from archivist_snapshotter.upload import S3Uploader
from tests.helpers import make_config


class RecordingClient:
    def __init__(self) -> None:
        self.uploads: list[tuple[str, str, str]] = []

    def upload_file(self, filename: str, bucket: str, key: str) -> object:
        self.uploads.append((filename, bucket, key))
        return None


class RecordingFactory:
    def __init__(self, client: RecordingClient) -> None:
        self.client = client
        self.calls: list[tuple[str, dict[str, object]]] = []

    def __call__(self, service_name: str, **kwargs: object) -> RecordingClient:
        self.calls.append((service_name, kwargs))
        return self.client


def test_upload_uses_explicit_s3_configuration(tmp_path: Path) -> None:
    config = make_config(tmp_path)
    archive = tmp_path / "archive.tar.gz"
    archive.write_bytes(b"archive")
    client = RecordingClient()
    factory = RecordingFactory(client)

    S3Uploader(config.s3, JsonLogger(io.StringIO()), client_factory=factory).upload(
        archive,
        "prod/archive.tar.gz",
    )

    assert factory.calls == [
        (
            "s3",
            {
                "endpoint_url": "https://s3.example.test",
                "region_name": "fr-par",
                "aws_access_key_id": "access-key",
                "aws_secret_access_key": "secret-key",
            },
        )
    ]
    assert client.uploads == [(str(archive), "archivist", "prod/archive.tar.gz")]


def test_upload_logs_do_not_include_credentials(tmp_path: Path) -> None:
    config = make_config(tmp_path)
    archive = tmp_path / "archive.tar.gz"
    archive.write_bytes(b"archive")
    logs = io.StringIO()

    S3Uploader(
        config.s3,
        JsonLogger(logs),
        client_factory=RecordingFactory(RecordingClient()),
    ).upload(archive, "prod/archive.tar.gz")

    log_text = logs.getvalue()
    assert "secret-key" not in log_text
    assert "access-key" not in log_text
    assert "upload_succeeded" in log_text
