from __future__ import annotations

import io
import json

from archivist_snapshotter.config import ConfigError
from archivist_snapshotter.logging import JsonLogger


def test_logger_includes_config_error_message() -> None:
    logs = io.StringIO()

    JsonLogger(logs).error(
        "config_invalid",
        error=ConfigError("ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY is required"),
    )

    payload = json.loads(logs.getvalue())
    assert payload["event"] == "config_invalid"
    assert payload["error"] == {
        "type": "ConfigError",
        "message": "ARCHIVIST_SNAPSHOTTER_S3_SECRET_ACCESS_KEY is required",
    }


def test_logger_redacts_secret_like_fields() -> None:
    logs = io.StringIO()
    redacted = "[redacted]"

    JsonLogger(logs).error(
        "config_invalid",
        access_key_id="access-key",
        secret_access_key="secret-key",  # noqa: S106 - inert redaction fixture
        token="token-value",  # noqa: S106 - inert redaction fixture
        nested={"password": "password-value", "bucket": "archivist"},
    )

    log_text = logs.getvalue()
    payload = json.loads(log_text)
    assert payload["access_key_id"] == redacted
    assert payload["secret_access_key"] == redacted
    assert payload["token"] == redacted
    assert payload["nested"] == {"bucket": "archivist", "password": redacted}
    assert "access-key" not in log_text
    assert "secret-key" not in log_text
    assert "token-value" not in log_text
    assert "password-value" not in log_text
