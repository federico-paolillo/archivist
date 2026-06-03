from __future__ import annotations

import os
from collections.abc import Mapping
from dataclasses import dataclass
from pathlib import Path

DEFAULT_DATA_DIR = Path("/data")
DEFAULT_SQLITE_PATH = Path("/data/archive.db")
DEFAULT_INTERVAL_SECONDS = 86_400
DEFAULT_WORK_DIR = Path("/tmp/archivist-snapshotter")  # noqa: S108 - canonical service work dir


class ConfigError(ValueError):
    pass


@dataclass(frozen=True)
class S3Config:
    endpoint_url: str
    region: str
    bucket: str
    access_key_id: str
    secret_access_key: str
    object_prefix: str = ""


@dataclass(frozen=True)
class Config:
    data_dir: Path
    sqlite_path: Path
    interval_seconds: int
    work_dir: Path
    s3: S3Config


def load_config(environ: Mapping[str, str] | None = None) -> Config:
    env = os.environ if environ is None else environ

    interval_seconds = _parse_positive_int(
        env.get("ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS"),
        default=DEFAULT_INTERVAL_SECONDS,
        name="ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS",
    )

    data_dir = Path(env.get("ARCHIVIST_DATA_DIR", str(DEFAULT_DATA_DIR)))
    sqlite_path = Path(env.get("ARCHIVIST_SQLITE_PATH", str(DEFAULT_SQLITE_PATH)))
    work_dir = Path(env.get("ARCHIVIST_SNAPSHOTTER_WORK_DIR", str(DEFAULT_WORK_DIR)))
    _reject_work_dir_inside_data_dir(data_dir=data_dir, work_dir=work_dir)

    return Config(
        data_dir=data_dir,
        sqlite_path=sqlite_path,
        interval_seconds=interval_seconds,
        work_dir=work_dir,
        s3=S3Config(
            endpoint_url=_required(env, "ARCHIVIST_SNAPSHOTTER_S3_ENDPOINT_URL"),
            region=_required(env, "ARCHIVIST_SNAPSHOTTER_S3_REGION"),
            bucket=_required(env, "ARCHIVIST_SNAPSHOTTER_S3_BUCKET"),
            access_key_id=_required(env, "ARCHIVIST_SNAPSHOTTER_S3_ACCESS_KEY_ID"),
            secret_access_key=_required(env, "ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY"),
            object_prefix=normalize_object_prefix(
                env.get("ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX", "")
            ),
        ),
    )


def normalize_object_prefix(value: str) -> str:
    return "/".join(part for part in value.strip().split("/") if part)


def _required(env: Mapping[str, str], name: str) -> str:
    value = env.get(name, "").strip()
    if value == "":
        raise ConfigError(f"{name} is required")
    return value


def _parse_positive_int(value: str | None, *, default: int, name: str) -> int:
    if value is None or value.strip() == "":
        return default
    try:
        parsed = int(value)
    except ValueError as exc:
        raise ConfigError(f"{name} must be a positive integer") from exc
    if parsed <= 0:
        raise ConfigError(f"{name} must be a positive integer")
    return parsed


def _reject_work_dir_inside_data_dir(*, data_dir: Path, work_dir: Path) -> None:
    resolved_data_dir = data_dir.expanduser().resolve(strict=False)
    resolved_work_dir = work_dir.expanduser().resolve(strict=False)
    if resolved_work_dir == resolved_data_dir or resolved_data_dir in resolved_work_dir.parents:
        raise ConfigError("ARCHIVIST_SNAPSHOTTER_WORK_DIR must not be inside ARCHIVIST_DATA_DIR")
