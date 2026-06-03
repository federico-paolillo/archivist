from __future__ import annotations

import argparse
import asyncio
import sys

from archivist_snapshotter.archive import create_snapshot_archive
from archivist_snapshotter.config import ConfigError, load_config
from archivist_snapshotter.daemon import run_daemon
from archivist_snapshotter.logging import JsonLogger
from archivist_snapshotter.telemetry import bootstrap_telemetry
from archivist_snapshotter.upload import S3Uploader


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        prog="archivist-snapshotter",
        description="Archive Archivist data and upload it to S3-compatible object storage.",
    )
    parser.parse_args(argv)

    logger = JsonLogger()
    telemetry = bootstrap_telemetry()

    try:
        try:
            config = load_config()
        except ConfigError as exc:
            logger.error("config_invalid", error=exc)
            return 2

        uploader = S3Uploader(config.s3, logger)

        try:
            asyncio.run(
                run_daemon(
                    config,
                    logger,
                    capture=lambda: create_snapshot_archive(config, logger),
                    upload=lambda snapshot: uploader.upload(
                        snapshot.archive_path,
                        snapshot.object_key,
                    ),
                )
            )
        except KeyboardInterrupt:
            logger.info("daemon_stopped", reason="keyboard_interrupt")
            return 0
    finally:
        telemetry.shutdown()

    return 0


if __name__ == "__main__":
    sys.exit(main())
