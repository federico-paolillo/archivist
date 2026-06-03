from __future__ import annotations

from typing import Any, ClassVar, cast

from opentelemetry.sdk._logs import LoggerProvider
from opentelemetry.sdk.resources import SERVICE_NAME
from opentelemetry.sdk.trace import TracerProvider

from archivist_snapshotter import telemetry


class RecordingTracerProvider:
    def __init__(self, *, resource: object) -> None:
        self.resource = resource
        self.processors: list[object] = []
        self.shutdown_called = False

    def add_span_processor(self, processor: object) -> None:
        self.processors.append(processor)

    def shutdown(self) -> None:
        self.shutdown_called = True


class RecordingLoggerProvider:
    def __init__(self, *, resource: object) -> None:
        self.resource = resource
        self.processors: list[object] = []
        self.shutdown_called = False

    def add_log_record_processor(self, processor: object) -> None:
        self.processors.append(processor)

    def shutdown(self) -> None:
        self.shutdown_called = True


class RecordingSpanExporter:
    pass


class RecordingLogExporter:
    pass


class RecordingSpanProcessor:
    def __init__(self, exporter: object) -> None:
        self.exporter = exporter


class RecordingLogProcessor:
    def __init__(self, exporter: object) -> None:
        self.exporter = exporter


class RecordingBotocoreInstrumentor:
    instrumented_providers: ClassVar[list[object]] = []

    def instrument(self, *, tracer_provider: object) -> None:
        self.instrumented_providers.append(tracer_provider)


def test_bootstrap_telemetry_ignores_removed_exporter_toggles(
    monkeypatch: Any,
) -> None:
    monkeypatch.setattr(telemetry, "_telemetry", None)
    monkeypatch.setenv("OTEL_TRACES_EXPORTER", "none")
    monkeypatch.setenv("OTEL_LOGS_EXPORTER", "none")

    trace_providers: list[object] = []
    log_providers: list[object] = []
    RecordingBotocoreInstrumentor.instrumented_providers = []

    def set_tracer_provider(provider: object) -> None:
        trace_providers.append(provider)

    def set_logger_provider(provider: object) -> None:
        log_providers.append(provider)

    monkeypatch.setattr(telemetry, "TracerProvider", RecordingTracerProvider)
    monkeypatch.setattr(telemetry, "LoggerProvider", RecordingLoggerProvider)
    monkeypatch.setattr(telemetry, "OTLPSpanExporter", RecordingSpanExporter)
    monkeypatch.setattr(telemetry, "OTLPLogExporter", RecordingLogExporter)
    monkeypatch.setattr(telemetry, "BatchSpanProcessor", RecordingSpanProcessor)
    monkeypatch.setattr(telemetry, "BatchLogRecordProcessor", RecordingLogProcessor)
    monkeypatch.setattr(telemetry, "BotocoreInstrumentor", RecordingBotocoreInstrumentor)
    monkeypatch.setattr(telemetry.trace, "set_tracer_provider", set_tracer_provider)
    monkeypatch.setattr(telemetry._logs, "set_logger_provider", set_logger_provider)

    configured = telemetry.bootstrap_telemetry()

    assert configured.botocore_instrumented is True
    assert trace_providers == [configured.trace_provider]
    assert log_providers == [configured.log_provider]
    assert RecordingBotocoreInstrumentor.instrumented_providers == [configured.trace_provider]

    assert isinstance(configured.trace_provider, RecordingTracerProvider)
    assert isinstance(configured.log_provider, RecordingLoggerProvider)
    assert configured.trace_provider.resource.attributes[SERVICE_NAME] == "archivist-snapshotter"
    assert configured.log_provider.resource.attributes[SERVICE_NAME] == "archivist-snapshotter"

    trace_processor = configured.trace_provider.processors[0]
    log_processor = configured.log_provider.processors[0]
    assert isinstance(trace_processor, RecordingSpanProcessor)
    assert isinstance(trace_processor.exporter, RecordingSpanExporter)
    assert isinstance(log_processor, RecordingLogProcessor)
    assert isinstance(log_processor.exporter, RecordingLogExporter)


def test_telemetry_shutdown_suppresses_provider_errors(monkeypatch: Any) -> None:
    class FailingProvider:
        def shutdown(self) -> None:
            raise RuntimeError("collector unavailable")

    configured = telemetry.Telemetry(
        trace_provider=cast(TracerProvider, FailingProvider()),
        log_provider=cast(LoggerProvider, FailingProvider()),
        botocore_instrumented=True,
    )

    configured.shutdown()
