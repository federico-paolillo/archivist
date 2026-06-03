from __future__ import annotations

import sqlite3
from pathlib import Path

from archivist_snapshotter.config import Config, S3Config


def make_config(
    tmp_path: Path,
    *,
    data_dir: Path | None = None,
    sqlite_path: Path | None = None,
    work_dir: Path | None = None,
    object_prefix: str = "prod",
    interval_seconds: int = 86_400,
    endpoint_url: str = "https://s3.example.test",
    region: str = "fr-par",
    bucket: str = "archivist",
    access_key_id: str = "access-key",
    secret_access_key: str = "secret-key",  # noqa: S107 - inert test fixture value
) -> Config:
    resolved_data_dir = tmp_path / "data" if data_dir is None else data_dir
    resolved_sqlite_path = resolved_data_dir / "archive.db" if sqlite_path is None else sqlite_path
    resolved_work_dir = tmp_path / "work" if work_dir is None else work_dir
    return Config(
        data_dir=resolved_data_dir,
        sqlite_path=resolved_sqlite_path,
        interval_seconds=interval_seconds,
        work_dir=resolved_work_dir,
        s3=S3Config(
            endpoint_url=endpoint_url,
            region=region,
            bucket=bucket,
            access_key_id=access_key_id,
            secret_access_key=secret_access_key,
            object_prefix=object_prefix,
        ),
    )


def create_sqlite_database(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with sqlite3.connect(path) as db:
        db.execute("CREATE TABLE articles (id TEXT PRIMARY KEY, title TEXT NOT NULL)")
        db.execute("INSERT INTO articles (id, title) VALUES ('01HX', 'stored article')")
