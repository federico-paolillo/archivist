from __future__ import annotations

import io
import json

from opentelemetry import trace
from opentelemetry.trace import NonRecordingSpan, SpanContext, TraceFlags, TraceState, use_span

from archivist_snapshotter import logging as logging_module
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


def test_logger_includes_active_span_ids() -> None:
    logs = io.StringIO()
    span_context = SpanContext(
        trace_id=0x1234567890ABCDEF1234567890ABCDEF,
        span_id=0x1234567890ABCDEF,
        is_remote=False,
        trace_flags=TraceFlags(TraceFlags.SAMPLED),
        trace_state=TraceState(),
    )

    with use_span(NonRecordingSpan(span_context)):
        JsonLogger(logs).info("inside_span")

    payload = json.loads(logs.getvalue())
    assert payload["trace_id"] == trace.format_trace_id(span_context.trace_id)
    assert payload["span_id"] == trace.format_span_id(span_context.span_id)


def test_debug_logger_writes_stdout_without_emitting_otel(monkeypatch) -> None:
    logs = io.StringIO()
    otel_calls = 0

    def get_logger(_name: str) -> object:
        nonlocal otel_calls
        otel_calls += 1
        raise AssertionError("debug logs must not touch the OTEL logger")

    monkeypatch.setattr(logging_module._logs, "get_logger", get_logger)

    JsonLogger(logs).debug("heartbeat", status="idle")

    payload = json.loads(logs.getvalue())
    assert payload["level"] == "debug"
    assert payload["event"] == "heartbeat"
    assert payload["status"] == "idle"
    assert otel_calls == 0
