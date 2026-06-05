from __future__ import annotations

import asyncio
from collections.abc import Awaitable, Callable

from archivist_snapshotter.archive import SnapshotResult
from archivist_snapshotter.config import Config
from archivist_snapshotter.logging import JsonLogger
from archivist_snapshotter.telemetry import (
    get_tracer,
    independent_root_context,
    mark_span_error,
    set_span_attributes,
)

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
        await run_once(logger, capture=capture, upload=upload)
        attempts += 1
        await sleep(config.interval_seconds)


async def run_once(logger: JsonLogger, *, capture: Capture, upload: Upload) -> None:
    with get_tracer().start_as_current_span(
        "snapshotter.run_once",
        context=independent_root_context(),
        record_exception=False,
        set_status_on_exception=False,
    ) as span:
        snapshot: SnapshotResult | None = None
        try:
            logger.info("snapshot_started")
            snapshot = capture()
            set_span_attributes(
                span,
                {
                    "archive.name": snapshot.archive_name,
                    "s3.object_key": snapshot.object_key,
                },
            )
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
            mark_span_error(span)
            logger.error("snapshot_failed", error=exc)
        finally:
            if snapshot is not None:
                try:
                    with get_tracer().start_as_current_span(
                        "snapshotter.cleanup",
                        attributes={"archive.name": snapshot.archive_name},
                        record_exception=False,
                        set_status_on_exception=False,
                    ) as cleanup_span:
                        try:
                            snapshot.cleanup()
                        except Exception:
                            mark_span_error(cleanup_span)
                            raise
                    logger.info("cleanup_succeeded", archive_name=snapshot.archive_name)
                except Exception as exc:
                    mark_span_error(span)
                    logger.error("cleanup_failed", archive_name=snapshot.archive_name, error=exc)
