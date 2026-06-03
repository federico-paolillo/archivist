from __future__ import annotations

import pytest

from archivist_snapshotter.config import (
    DEFAULT_DATA_DIR,
    DEFAULT_INTERVAL_SECONDS,
    DEFAULT_SQLITE_PATH,
    DEFAULT_WORK_DIR,
    ConfigError,
    load_config,
    normalize_object_prefix,
)


def required_env() -> dict[str, str]:
    return {
        "ARCHIVIST_SNAPSHOTTER_S3_ENDPOINT_URL": "https://s3.fr-par.scw.cloud",
        "ARCHIVIST_SNAPSHOTTER_S3_REGION": "fr-par",
        "ARCHIVIST_SNAPSHOTTER_S3_BUCKET": "archivist-backups",
        "ARCHIVIST_SNAPSHOTTER_S3_ACCESS_KEY_ID": "access",
        "ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY": "secret",
    }


def test_load_config_uses_defaults() -> None:
    config = load_config(required_env())

    assert config.data_dir == DEFAULT_DATA_DIR
    assert config.sqlite_path == DEFAULT_SQLITE_PATH
    assert config.interval_seconds == DEFAULT_INTERVAL_SECONDS
    assert config.work_dir == DEFAULT_WORK_DIR
    assert config.s3.object_prefix == ""


def test_load_config_reads_explicit_values() -> None:
    env = required_env()
    env.update(
        {
            "ARCHIVIST_DATA_DIR": "/mnt/data",
            "ARCHIVIST_SQLITE_PATH": "/mnt/data/custom.db",
            "ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS": "42",
            "ARCHIVIST_SNAPSHOTTER_WORK_DIR": "/mnt/work",
            "ARCHIVIST_SNAPSHOTTER_OBJECT_PREFIX": "/prod//daily/",
        }
    )

    config = load_config(env)

    assert str(config.data_dir) == "/mnt/data"
    assert str(config.sqlite_path) == "/mnt/data/custom.db"
    assert config.interval_seconds == 42
    assert str(config.work_dir) == "/mnt/work"
    assert config.s3.object_prefix == "prod/daily"


@pytest.mark.parametrize("value", ["0", "-1", "not-int"])
def test_load_config_rejects_invalid_interval(value: str) -> None:
    env = required_env()
    env["ARCHIVIST_SNAPSHOTTER_INTERVAL_SECONDS"] = value

    with pytest.raises(ConfigError):
        load_config(env)


def test_load_config_requires_s3_values() -> None:
    env = required_env()
    env["ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY"] = ""

    with pytest.raises(ConfigError, match="ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY"):
        load_config(env)


def test_load_config_rejects_work_dir_inside_data_dir() -> None:
    env = required_env()
    env["ARCHIVIST_DATA_DIR"] = "/data"
    env["ARCHIVIST_SNAPSHOTTER_WORK_DIR"] = "/data/.snapshotter"

    with pytest.raises(ConfigError, match="WORK_DIR must not be inside"):
        load_config(env)


def test_normalize_object_prefix() -> None:
    assert normalize_object_prefix("") == ""
    assert normalize_object_prefix("prod") == "prod"
    assert normalize_object_prefix("/prod//daily/") == "prod/daily"
