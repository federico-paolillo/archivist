from __future__ import annotations

from collections.abc import Mapping
from contextlib import suppress
from dataclasses import dataclass

from opentelemetry import _logs, trace
from opentelemetry.context import Context
from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.botocore import BotocoreInstrumentor
from opentelemetry.sdk._logs import LoggerProvider
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.trace import INVALID_SPAN, Status, StatusCode, Tracer, set_span_in_context
from opentelemetry.trace.span import Span
from opentelemetry.util.types import AttributeValue

from archivist_snapshotter import __version__

SNAPSHOTTER_SERVICE_NAME = "archivist-snapshotter"
INSTRUMENTATION_NAME = "archivist_snapshotter"

_telemetry: Telemetry | None = None


@dataclass(frozen=True)
class Telemetry:
    trace_provider: TracerProvider
    log_provider: LoggerProvider
    botocore_instrumented: bool

    def shutdown(self) -> None:
        with suppress(Exception):
            self.log_provider.shutdown()
        with suppress(Exception):
            self.trace_provider.shutdown()


def bootstrap_telemetry() -> Telemetry:
    global _telemetry

    if _telemetry is not None:
        return _telemetry

    resource = Resource.create({SERVICE_NAME: SNAPSHOTTER_SERVICE_NAME})
    trace_provider = _configure_traces(resource)
    log_provider = _configure_logs(resource)
    BotocoreInstrumentor().instrument(tracer_provider=trace_provider)

    _telemetry = Telemetry(
        trace_provider=trace_provider,
        log_provider=log_provider,
        botocore_instrumented=True,
    )
    return _telemetry


def get_tracer() -> Tracer:
    return trace.get_tracer(INSTRUMENTATION_NAME, __version__)


def independent_root_context() -> Context:
    return set_span_in_context(INVALID_SPAN)


def active_span_ids() -> dict[str, str]:
    span_context = trace.get_current_span().get_span_context()
    if not span_context.is_valid:
        return {}
    return {
        "trace_id": trace.format_trace_id(span_context.trace_id),
        "span_id": trace.format_span_id(span_context.span_id),
    }


def set_span_attributes(span: Span, attributes: Mapping[str, object]) -> None:
    for key, value in attributes.items():
        if value is None:
            continue
        span.set_attribute(key, _span_attribute(value))


def mark_span_error(span: Span) -> None:
    span.set_status(Status(StatusCode.ERROR))


def _span_attribute(value: object) -> AttributeValue:
    if isinstance(value, (str, bool, int, float)):
        return value
    return str(value)


def _configure_traces(resource: Resource) -> TracerProvider:
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter()))
    trace.set_tracer_provider(provider)
    return provider


def _configure_logs(resource: Resource) -> LoggerProvider:
    provider = LoggerProvider(resource=resource)
    provider.add_log_record_processor(BatchLogRecordProcessor(OTLPLogExporter()))
    _logs.set_logger_provider(provider)
    return provider
