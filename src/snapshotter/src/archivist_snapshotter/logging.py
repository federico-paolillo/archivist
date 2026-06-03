from __future__ import annotations

import json
import sys
from collections.abc import Mapping
from contextlib import suppress
from datetime import UTC, datetime
from typing import TextIO

from opentelemetry import _logs
from opentelemetry._logs import SeverityNumber
from opentelemetry.util.types import AnyValue

from archivist_snapshotter.config import ConfigError
from archivist_snapshotter.telemetry import INSTRUMENTATION_NAME, active_span_ids

SECRET_FIELD_FRAGMENTS = ("secret", "access_key", "credential", "token", "password")
SAFE_EXCEPTION_MESSAGE_TYPES = (ConfigError,)
OTEL_SEVERITY_BY_LEVEL = {
    "info": SeverityNumber.INFO,
    "error": SeverityNumber.ERROR,
}


class JsonLogger:
    def __init__(self, stream: TextIO | None = None) -> None:
        self._stream = sys.stdout if stream is None else stream

    def info(self, event: str, **fields: object) -> None:
        self._write("info", event, fields)

    def error(self, event: str, **fields: object) -> None:
        self._write("error", event, fields)

    def _write(self, level: str, event: str, fields: Mapping[str, object]) -> None:
        payload: dict[str, object] = {
            "level": level,
            "event": event,
            "time": datetime.now(UTC).isoformat().replace("+00:00", "Z"),
        }
        payload.update(_sanitize_fields(fields))
        payload.update(active_span_ids())
        self._stream.write(json.dumps(payload, sort_keys=True, separators=(",", ":")) + "\n")
        self._stream.flush()
        _emit_otel_log(level, event, payload)


def _sanitize_fields(fields: Mapping[str, object]) -> dict[str, object]:
    return {key: _sanitize_value(key, value) for key, value in fields.items()}


def _sanitize_value(key: str, value: object) -> object:
    if _is_secret_field(key):
        return "[redacted]"
    if isinstance(value, Mapping):
        return {
            str(nested_key): _sanitize_value(str(nested_key), nested_value)
            for nested_key, nested_value in value.items()
        }
    if isinstance(value, (list, tuple)):
        return [_sanitize_value(key, item) for item in value]
    if isinstance(value, Exception):
        exception: dict[str, object] = {"type": type(value).__name__}
        if isinstance(value, SAFE_EXCEPTION_MESSAGE_TYPES):
            exception["message"] = str(value)
        return exception
    return _json_safe(value)


def _is_secret_field(key: str) -> bool:
    normalized = key.lower()
    return any(fragment in normalized for fragment in SECRET_FIELD_FRAGMENTS)


def _json_safe(value: object) -> object:
    try:
        json.dumps(value)
    except TypeError:
        return str(value)
    return value


def _emit_otel_log(level: str, event: str, payload: Mapping[str, object]) -> None:
    with suppress(Exception):
        _logs.get_logger(INSTRUMENTATION_NAME).emit(
            severity_number=OTEL_SEVERITY_BY_LEVEL[level],
            severity_text=level.upper(),
            body=event,
            attributes=_otel_attributes(payload),
            event_name=event,
        )


def _otel_attributes(payload: Mapping[str, object]) -> dict[str, AnyValue]:
    return {key: _otel_attribute_value(value) for key, value in payload.items()}


def _otel_attribute_value(value: object) -> AnyValue:
    if value is None or isinstance(value, (str, bool, int, float, bytes)):
        return value
    if isinstance(value, Mapping):
        return {str(key): _otel_attribute_value(nested) for key, nested in value.items()}
    if isinstance(value, (list, tuple)):
        return [_otel_attribute_value(item) for item in value]
    return str(value)
