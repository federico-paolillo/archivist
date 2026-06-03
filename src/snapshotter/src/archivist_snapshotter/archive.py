from __future__ import annotations

import json
import shutil
import sqlite3
import tarfile
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path

from archivist_snapshotter import __version__
from archivist_snapshotter.config import Config, normalize_object_prefix
from archivist_snapshotter.logging import JsonLogger
from archivist_snapshotter.telemetry import get_tracer, mark_span_error, set_span_attributes

CONSISTENCY_NOTE = (
    "SQLite captured with the SQLite online backup API; non-database artifacts copied "
    "best-effort from a live filesystem without pausing Archivist services."
)


@dataclass(frozen=True)
class SnapshotTimestamp:
    created_at: datetime
    unix_timestamp: int
    archive_name: str


@dataclass(frozen=True)
class SnapshotResult:
    archive_name: str
    object_key: str
    archive_path: Path
    attempt_dir: Path

    def cleanup(self) -> None:
        shutil.rmtree(self.attempt_dir, ignore_errors=False)


@dataclass(frozen=True)
class ArtifactCopyStats:
    copied_files: int = 0
    skipped_files: int = 0


def snapshot_timestamp(created_at: datetime | None = None) -> SnapshotTimestamp:
    instant = datetime.now(UTC) if created_at is None else _as_utc(created_at)
    unix_timestamp = int(instant.timestamp())
    archive_name = f"archivist-{instant.strftime('%Y-%m-%d')}-{unix_timestamp}.tar.gz"
    return SnapshotTimestamp(
        created_at=instant,
        unix_timestamp=unix_timestamp,
        archive_name=archive_name,
    )


def object_key_for_archive(archive_name: str, object_prefix: str) -> str:
    prefix = normalize_object_prefix(object_prefix)
    if prefix == "":
        return archive_name
    return f"{prefix}/{archive_name}"


def create_snapshot_archive(
    config: Config,
    logger: JsonLogger,
    *,
    created_at: datetime | None = None,
) -> SnapshotResult:
    timestamp = snapshot_timestamp(created_at)
    object_key = object_key_for_archive(timestamp.archive_name, config.s3.object_prefix)
    attempt_dir = config.work_dir / timestamp.archive_name.removesuffix(".tar.gz")
    stage_root = attempt_dir / "stage"
    staged_data_dir = stage_root / "data"
    archive_path = attempt_dir / timestamp.archive_name

    if attempt_dir.exists():
        shutil.rmtree(attempt_dir)

    with get_tracer().start_as_current_span(
        "snapshotter.archive.create",
        attributes={
            "archive.name": timestamp.archive_name,
            "s3.object_key": object_key,
            "snapshotter.work_dir": str(config.work_dir),
        },
        record_exception=False,
        set_status_on_exception=False,
    ) as span:
        try:
            staged_data_dir.mkdir(parents=True)
            with get_tracer().start_as_current_span(
                "snapshotter.artifact_copy",
                attributes={
                    "source.data_dir": str(config.data_dir),
                    "target.data_dir": str(staged_data_dir),
                },
                record_exception=False,
                set_status_on_exception=False,
            ) as copy_span:
                try:
                    copy_stats = _copy_artifacts_best_effort(
                        data_dir=config.data_dir,
                        sqlite_path=config.sqlite_path,
                        staged_data_dir=staged_data_dir,
                        logger=logger,
                    )
                except Exception:
                    mark_span_error(copy_span)
                    raise
                set_span_attributes(
                    copy_span,
                    {
                        "artifact.copied_files": copy_stats.copied_files,
                        "artifact.skipped_files": copy_stats.skipped_files,
                    },
                )
            with get_tracer().start_as_current_span(
                "snapshotter.sqlite_backup",
                attributes={
                    "source.sqlite_path": str(config.sqlite_path),
                    "target.sqlite_path": str(_staged_sqlite_path(config, staged_data_dir)),
                },
                record_exception=False,
                set_status_on_exception=False,
            ) as backup_span:
                try:
                    _backup_sqlite(config.sqlite_path, _staged_sqlite_path(config, staged_data_dir))
                except Exception:
                    mark_span_error(backup_span)
                    raise
            with get_tracer().start_as_current_span(
                "snapshotter.manifest_write",
                attributes={"manifest.path": str(stage_root / "manifest.json")},
                record_exception=False,
                set_status_on_exception=False,
            ) as manifest_span:
                try:
                    _write_manifest(
                        stage_root=stage_root,
                        timestamp=timestamp,
                        object_key=object_key,
                        config=config,
                    )
                except Exception:
                    mark_span_error(manifest_span)
                    raise
            with get_tracer().start_as_current_span(
                "snapshotter.tarball_create",
                attributes={"archive.path": str(archive_path)},
                record_exception=False,
                set_status_on_exception=False,
            ) as tarball_span:
                try:
                    _create_tarball(stage_root=stage_root, archive_path=archive_path)
                except Exception:
                    mark_span_error(tarball_span)
                    raise
            set_span_attributes(span, {"archive.path": str(archive_path)})
        except Exception:
            mark_span_error(span)
            shutil.rmtree(attempt_dir, ignore_errors=True)
            raise

    logger.info(
        "archive_created",
        archive_name=timestamp.archive_name,
        object_key=object_key,
        archive_path=str(archive_path),
    )
    return SnapshotResult(
        archive_name=timestamp.archive_name,
        object_key=object_key,
        archive_path=archive_path,
        attempt_dir=attempt_dir,
    )


def _copy_artifacts_best_effort(
    *,
    data_dir: Path,
    sqlite_path: Path,
    staged_data_dir: Path,
    logger: JsonLogger,
) -> ArtifactCopyStats:
    sqlite_excluded_paths = _sqlite_excluded_paths(sqlite_path)
    copied_files = 0
    skipped_files = 0
    for source in data_dir.rglob("*"):
        if source in sqlite_excluded_paths:
            continue
        try:
            relative = source.relative_to(data_dir)
        except ValueError:
            continue

        target = staged_data_dir / relative
        try:
            if source.is_dir():
                target.mkdir(parents=True, exist_ok=True)
                continue
            if source.is_symlink():
                logger.info("artifact_skipped", reason="symlink", path=str(relative))
                skipped_files += 1
                continue
            if not source.is_file():
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
            copied_files += 1
        except FileNotFoundError:
            logger.info("artifact_skipped", reason="disappeared", path=str(relative))
            skipped_files += 1
    return ArtifactCopyStats(copied_files=copied_files, skipped_files=skipped_files)


def _backup_sqlite(source_path: Path, target_path: Path) -> None:
    target_path.parent.mkdir(parents=True, exist_ok=True)
    with (
        sqlite3.connect(f"file:{source_path}?mode=ro", uri=True) as source,
        sqlite3.connect(target_path) as target,
    ):
        source.backup(target)


def _write_manifest(
    *,
    stage_root: Path,
    timestamp: SnapshotTimestamp,
    object_key: str,
    config: Config,
) -> None:
    manifest = {
        "archive_name": timestamp.archive_name,
        "object_key": object_key,
        "created_at": timestamp.created_at.isoformat().replace("+00:00", "Z"),
        "unix_timestamp": timestamp.unix_timestamp,
        "source_data_dir": str(config.data_dir),
        "source_sqlite_path": str(config.sqlite_path),
        "snapshotter_version": __version__,
        "consistency": CONSISTENCY_NOTE,
    }
    (stage_root / "manifest.json").write_text(
        json.dumps(manifest, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )


def _create_tarball(*, stage_root: Path, archive_path: Path) -> None:
    with tarfile.open(archive_path, "w:gz") as archive:
        archive.add(stage_root / "manifest.json", arcname="manifest.json")
        archive.add(stage_root / "data", arcname="data")


def _sqlite_excluded_paths(sqlite_path: Path) -> set[Path]:
    return {
        sqlite_path,
        Path(f"{sqlite_path}-wal"),
        Path(f"{sqlite_path}-shm"),
        Path(f"{sqlite_path}-journal"),
    }


def _staged_sqlite_path(config: Config, staged_data_dir: Path) -> Path:
    try:
        relative = config.sqlite_path.relative_to(config.data_dir)
    except ValueError:
        relative = Path(config.sqlite_path.name)
    return staged_data_dir / relative


def _as_utc(value: datetime) -> datetime:
    if value.tzinfo is None:
        return value.replace(tzinfo=UTC)
    return value.astimezone(UTC)
