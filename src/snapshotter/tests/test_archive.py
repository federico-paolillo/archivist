from __future__ import annotations

import io
import json
import sqlite3
import tarfile
from datetime import UTC, datetime
from pathlib import Path

import pytest

from archivist_snapshotter.archive import (
    CONSISTENCY_NOTE,
    create_snapshot_archive,
    object_key_for_archive,
    snapshot_timestamp,
)
from archivist_snapshotter.logging import JsonLogger
from tests.helpers import create_sqlite_database, make_config


def test_archive_name_and_object_key_are_deterministic() -> None:
    timestamp = snapshot_timestamp(datetime(2026, 6, 2, 10, 30, tzinfo=UTC))

    assert timestamp.archive_name == "archivist-2026-06-02-1780396200.tar.gz"
    assert object_key_for_archive(timestamp.archive_name, "prod") == (
        "prod/archivist-2026-06-02-1780396200.tar.gz"
    )


def test_create_snapshot_archive_uses_sqlite_backup_and_excludes_sidecars(tmp_path: Path) -> None:
    config = make_config(tmp_path)
    create_sqlite_database(config.sqlite_path)
    (config.data_dir / "archive.db-wal").write_text("live wal", encoding="utf-8")
    (config.data_dir / "archive.db-shm").write_text("live shm", encoding="utf-8")
    (config.data_dir / "articles" / "01HX").mkdir(parents=True)
    (config.data_dir / "articles" / "01HX" / "summary.md").write_text(
        "summary",
        encoding="utf-8",
    )

    result = create_snapshot_archive(
        config,
        JsonLogger(io.StringIO()),
        created_at=datetime(2026, 6, 2, 10, 30, tzinfo=UTC),
    )

    with tarfile.open(result.archive_path, "r:gz") as archive:
        names = set(archive.getnames())
        manifest = _read_json_member(archive, "manifest.json")
        backup_bytes = archive.extractfile("data/archive.db")
        assert backup_bytes is not None
        backup_path = tmp_path / "backup.db"
        backup_path.write_bytes(backup_bytes.read())

    assert "manifest.json" in names
    assert "data/archive.db" in names
    assert "data/archive.db-wal" not in names
    assert "data/archive.db-shm" not in names
    assert "data/articles/01HX/summary.md" in names
    assert manifest == {
        "archive_name": "archivist-2026-06-02-1780396200.tar.gz",
        "object_key": "prod/archivist-2026-06-02-1780396200.tar.gz",
        "created_at": "2026-06-02T10:30:00Z",
        "unix_timestamp": 1780396200,
        "source_data_dir": str(config.data_dir),
        "source_sqlite_path": str(config.sqlite_path),
        "snapshotter_version": "0.1.0",
        "consistency": CONSISTENCY_NOTE,
    }

    with sqlite3.connect(backup_path) as db:
        assert db.execute("SELECT title FROM articles WHERE id = '01HX'").fetchone() == (
            "stored article",
        )

    result.cleanup()
    assert not result.attempt_dir.exists()


def test_create_snapshot_archive_logs_and_skips_disappearing_file(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    config = make_config(tmp_path)
    create_sqlite_database(config.sqlite_path)
    disappearing = config.data_dir / "articles" / "missing.md"
    disappearing.parent.mkdir(parents=True)
    disappearing.write_text("gone", encoding="utf-8")
    original_copy2 = __import__("shutil").copy2

    def copy2_with_disappearance(source: str | Path, target: str | Path) -> object:
        if Path(source) == disappearing:
            disappearing.unlink()
        return original_copy2(source, target)

    monkeypatch.setattr("shutil.copy2", copy2_with_disappearance)
    logs = io.StringIO()

    result = create_snapshot_archive(config, JsonLogger(logs))

    assert '"event":"artifact_skipped"' in logs.getvalue()
    assert result.archive_path.exists()
    result.cleanup()


def test_create_snapshot_archive_skips_symlinks(tmp_path: Path) -> None:
    config = make_config(tmp_path)
    create_sqlite_database(config.sqlite_path)
    outside = tmp_path / "outside-secret.txt"
    outside.write_text("secret", encoding="utf-8")
    symlink = config.data_dir / "articles" / "linked-secret.txt"
    symlink.parent.mkdir(parents=True)
    symlink.symlink_to(outside)
    logs = io.StringIO()

    result = create_snapshot_archive(config, JsonLogger(logs))

    with tarfile.open(result.archive_path, "r:gz") as archive:
        names = set(archive.getnames())

    assert "data/articles/linked-secret.txt" not in names
    assert '"reason":"symlink"' in logs.getvalue()
    result.cleanup()


def _read_json_member(archive: tarfile.TarFile, name: str) -> dict[str, object]:
    member = archive.extractfile(name)
    assert member is not None
    return json.loads(member.read().decode("utf-8"))
