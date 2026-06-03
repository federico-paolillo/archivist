from __future__ import annotations

import asyncio
from collections.abc import Awaitable, Callable

from archivist_snapshotter.archive import SnapshotResult
from archivist_snapshotter.config import Config
from archivist_snapshotter.logging import JsonLogger

Sleep = Callable[[float], Awaitable[object]]
Capture = Callable[[], SnapshotResult]
Upload = Callable[[SnapshotResult], None]


async def run_daemon(
    config: Config,
    logger: JsonLogger,
    *,
    capture: Capture,
    upload: Upload,
    sleep: Sleep = asyncio.sleep,
    max_attempts: int | None = None,
) -> None:
    logger.info("daemon_started", interval_seconds=config.interval_seconds)
    attempts = 0
    while max_attempts is None or attempts < max_attempts:
        logger.info("daemon_sleeping", interval_seconds=config.interval_seconds)
        await sleep(config.interval_seconds)
        attempts += 1
        await run_once(logger, capture=capture, upload=upload)


async def run_once(logger: JsonLogger, *, capture: Capture, upload: Upload) -> None:
    snapshot: SnapshotResult | None = None
    try:
        logger.info("snapshot_started")
        snapshot = capture()
        try:
            upload(snapshot)
        except Exception as exc:
            logger.error(
                "upload_failed",
                archive_name=snapshot.archive_name,
                object_key=snapshot.object_key,
                error=exc,
            )
            raise
    except Exception as exc:
        logger.error("snapshot_failed", error=exc)
    finally:
        if snapshot is not None:
            try:
                snapshot.cleanup()
                logger.info("cleanup_succeeded", archive_name=snapshot.archive_name)
            except Exception as exc:
                logger.error("cleanup_failed", archive_name=snapshot.archive_name, error=exc)
