from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path


@dataclass(frozen=True, slots=True)
class Event:
    timestamp: datetime | None
    level: str
    service: str
    message: str
    correlation_id: str | None
    line: int


@dataclass(slots=True)
class Group:
    fingerprint: str
    count: int = 0
    first_line: int = 0
    last_line: int = 0
    first_timestamp: datetime | None = None
    last_timestamp: datetime | None = None
    example: str = ""


@dataclass(slots=True)
class Analysis:
    path: Path
    lines: int = 0
    malformed_lines: int = 0
    groups: dict[str, Group] = field(default_factory=dict)
    bursts: list[dict[str, object]] = field(default_factory=list)
    correlations: dict[str, list[dict[str, object]]] = field(default_factory=dict)
    privacy_counts: dict[str, int] = field(default_factory=dict)
