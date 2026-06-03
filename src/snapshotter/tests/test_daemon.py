from __future__ import annotations

import asyncio
import io
from pathlib import Path

from archivist_snapshotter.archive import SnapshotResult
from archivist_snapshotter.daemon import run_daemon
from archivist_snapshotter.logging import JsonLogger
from tests.helpers import make_config


def test_daemon_sleeps_first_and_continues_after_failure(tmp_path: Path) -> None:
    config = make_config(tmp_path, interval_seconds=7)
    slept: list[float] = []
    attempts = 0
    uploads: list[str] = []
    logs = io.StringIO()

    async def sleep(seconds: float) -> None:
        slept.append(seconds)

    def capture() -> SnapshotResult:
        nonlocal attempts
        attempts += 1
        return SnapshotResult(
            archive_name=f"archive-{attempts}.tar.gz",
            object_key=f"prod/archive-{attempts}.tar.gz",
            archive_path=tmp_path / f"archive-{attempts}.tar.gz",
            attempt_dir=tmp_path / f"attempt-{attempts}",
        )

    def upload(snapshot: SnapshotResult) -> None:
        uploads.append(snapshot.object_key)
        if len(uploads) == 1:
            raise RuntimeError("endpoint rejected upload")

    asyncio.run(
        run_daemon(
            config,
            JsonLogger(logs),
            capture=capture,
            upload=upload,
            sleep=sleep,
            max_attempts=2,
        )
    )

    assert slept == [7, 7]
    assert uploads == ["prod/archive-1.tar.gz", "prod/archive-2.tar.gz"]
    log_text = logs.getvalue()
    assert '"event":"upload_failed"' in log_text
    assert '"archive_name":"archive-1.tar.gz"' in log_text
    assert '"object_key":"prod/archive-1.tar.gz"' in log_text
    assert '"event":"snapshot_failed"' in log_text
    assert "endpoint rejected upload" not in log_text
