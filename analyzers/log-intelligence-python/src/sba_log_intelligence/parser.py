from __future__ import annotations

import json
import re
from datetime import UTC, datetime

from .models import Event

TIMESTAMP = re.compile(r"\d{4}-\d{2}-\d{2}[T ][0-9:.]+(?:Z|[+-][0-9:]+)?")
LEVEL = re.compile(r"\b(DEBUG|INFO|WARN(?:ING)?|ERROR|CRITICAL|FATAL)\b", re.IGNORECASE)
SERVICE = re.compile(r"\b(?:service|component)[=:]\s*([A-Za-z0-9._-]+)", re.IGNORECASE)
CORRELATION = re.compile(
    r"\b(?:request[_-]?id|trace[_-]?id|correlation[_-]?id)[=:]\s*([A-Za-z0-9._:-]+)",
    re.IGNORECASE,
)


def parse_timestamp(value: str) -> datetime | None:
    normalized = value.strip().replace(" ", "T", 1)
    if normalized.endswith("Z"):
        normalized = f"{normalized[:-1]}+00:00"
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=UTC)
    return parsed.astimezone(UTC)


def parse_line(raw: str, line_number: int) -> Event:
    timestamp: datetime | None = None
    level = "INFO"
    service = "unknown"
    message = raw.rstrip("\r\n")
    correlation_id: str | None = None
    try:
        value = json.loads(message)
    except json.JSONDecodeError:
        value = None
    if isinstance(value, dict):
        timestamp = parse_timestamp(_first_text(value, "timestamp", "time", "@timestamp", "ts"))
        level = _normalize_level(_first_text(value, "level", "severity"))
        service = _first_text(value, "service", "component", "logger") or "unknown"
        message = _first_text(value, "message", "msg", "error") or message
        correlation_id = (
            _first_text(value, "request_id", "requestId", "trace_id", "traceId", "correlation_id")
            or None
        )
    else:
        timestamp_match = TIMESTAMP.search(message)
        level_match = LEVEL.search(message)
        service_match = SERVICE.search(message)
        correlation_match = CORRELATION.search(message)
        if timestamp_match:
            timestamp = parse_timestamp(timestamp_match.group(0))
        if level_match:
            level = _normalize_level(level_match.group(1))
        if service_match:
            service = service_match.group(1)
        if correlation_match:
            correlation_id = correlation_match.group(1)
    return Event(timestamp, level, service, message, correlation_id, line_number)


def _first_text(value: dict[str, object], *keys: str) -> str:
    for key in keys:
        candidate = value.get(key)
        if isinstance(candidate, str):
            return candidate
    return ""


def _normalize_level(value: str) -> str:
    upper = value.upper()
    if upper in {"WARN", "WARNING"}:
        return "WARNING"
    if upper in {"FATAL", "CRITICAL"}:
        return "CRITICAL"
    return upper if upper in {"DEBUG", "INFO", "ERROR"} else "INFO"
