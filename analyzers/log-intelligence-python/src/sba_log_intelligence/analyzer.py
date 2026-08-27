from __future__ import annotations

import gzip
import re
from collections import Counter, deque
from datetime import timedelta
from pathlib import Path
from typing import TextIO

from .models import Analysis, Event, Group
from .parser import parse_line

UUID = re.compile(r"\b[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}\b", re.IGNORECASE)
IP = re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b")
NUMBER = re.compile(r"\b\d{3,}\b")
TIMESTAMP = re.compile(r"\d{4}-\d{2}-\d{2}[T ][0-9:.]+(?:Z|[+-][0-9:]+)?")
CORRELATION = re.compile(
    r"\b(request[_-]?id|trace[_-]?id|correlation[_-]?id)([=:]\s*)[A-Za-z0-9._:-]+",
    re.IGNORECASE,
)
SENSITIVE = {
    "email": re.compile(r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"),
    "bearer_token": re.compile(r"Bearer\s+[A-Za-z0-9._~+/=-]{12,}", re.IGNORECASE),
    "password": re.compile(r"\b(?:password|secret|api[_-]?key)\s*[:=]\s*\S+", re.IGNORECASE),
}
ERROR_LEVELS = {"ERROR", "CRITICAL"}


def normalize_message(message: str) -> str:
    value = TIMESTAMP.sub("<TIMESTAMP>", message)
    value = CORRELATION.sub(r"\1\2<ID>", value)
    value = UUID.sub("<UUID>", value)
    value = IP.sub("<IP>", value)
    value = NUMBER.sub("<NUMBER>", value)
    return " ".join(value.split())[:500]


def analyze_file(path: Path, burst_count: int = 5, burst_window_seconds: int = 60) -> Analysis:
    result = Analysis(path=path)
    recent_errors: deque[Event] = deque()
    privacy = Counter[str]()
    with _open_text(path) as stream:
        for line_number, raw in enumerate(stream, start=1):
            result.lines += 1
            if "\x00" in raw:
                result.malformed_lines += 1
                continue
            event = parse_line(raw, line_number)
            for kind, pattern in SENSITIVE.items():
                privacy[kind] += len(pattern.findall(raw))
            if event.correlation_id:
                chain = result.correlations.setdefault(event.correlation_id, [])
                if len(chain) < 100:
                    chain.append(_event_reference(event))
            if event.level not in ERROR_LEVELS:
                continue
            fingerprint = normalize_message(event.message)
            group = result.groups.setdefault(
                fingerprint,
                Group(fingerprint=fingerprint, first_line=line_number, example=event.message[:300]),
            )
            group.count += 1
            group.last_line = line_number
            group.first_timestamp = group.first_timestamp or event.timestamp
            group.last_timestamp = event.timestamp or group.last_timestamp
            if event.timestamp:
                recent_errors.append(event)
                cutoff = event.timestamp - timedelta(seconds=burst_window_seconds)
                while recent_errors:
                    oldest_timestamp = recent_errors[0].timestamp
                    if oldest_timestamp is None or oldest_timestamp >= cutoff:
                        break
                    recent_errors.popleft()
                if len(recent_errors) == burst_count:
                    result.bursts.append(
                        {
                            "startedAt": (
                                recent_errors[0].timestamp.isoformat()
                                if recent_errors[0].timestamp is not None
                                else event.timestamp.isoformat()
                            ),
                            "endedAt": event.timestamp.isoformat(),
                            "count": len(recent_errors),
                            "windowSeconds": burst_window_seconds,
                        }
                    )
    result.privacy_counts = dict(sorted((kind, count) for kind, count in privacy.items() if count))
    return result


def findings(result: Analysis) -> list[dict[str, object]]:
    output: list[dict[str, object]] = []
    groups = sorted(result.groups.values(), key=lambda item: (-item.count, item.fingerprint))
    for group in groups[:1000]:
        output.append(
            {
                "type": "finding",
                "finding": {
                    "ruleId": "PY_REPEATED_ERROR",
                    "severity": "medium" if group.count < 5 else "high",
                    "title": "Repeated log error pattern",
                    "summary": (
                        f"Observed {group.count} occurrence(s) of a normalized error pattern."
                    ),
                    "confidence": "strong",
                    "evidence": [
                        {
                            "artifact": str(result.path),
                            "lineStart": group.first_line,
                            "lineEnd": group.last_line,
                            "excerpt": group.example,
                        }
                    ],
                },
            }
        )
    for burst in result.bursts:
        output.append(
            {
                "type": "finding",
                "finding": {
                    "ruleId": "PY_ERROR_BURST",
                    "severity": "high",
                    "title": "Burst of error events",
                    "summary": (
                        f"At least {burst['count']} errors occurred within "
                        f"{burst['windowSeconds']} seconds."
                    ),
                    "confidence": "moderate",
                    "evidence": [],
                },
            }
        )
    return output


def _open_text(path: Path) -> TextIO:
    if path.suffix.lower() == ".gz":
        return gzip.open(path, mode="rt", encoding="utf-8", errors="replace")
    return path.open(mode="rt", encoding="utf-8", errors="replace")


def _event_reference(event: Event) -> dict[str, object]:
    return {
        "timestamp": event.timestamp.isoformat() if event.timestamp else None,
        "service": event.service,
        "level": event.level,
        "line": event.line,
    }
