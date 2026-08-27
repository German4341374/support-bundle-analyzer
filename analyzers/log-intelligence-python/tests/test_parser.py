from datetime import UTC

from sba_log_intelligence.parser import parse_line, parse_timestamp


def test_json_log() -> None:
    event = parse_line(
        '{"timestamp":"2026-01-01T01:02:03+02:00","level":"warn","service":"worker","message":"slow","trace_id":"t-1"}',
        4,
    )
    assert event.timestamp is not None and event.timestamp.tzinfo == UTC
    assert event.level == "WARNING"
    assert event.service == "worker"
    assert event.correlation_id == "t-1"


def test_generic_log() -> None:
    event = parse_line("2026-01-01T01:02:03Z CRITICAL service=api request_id=r-1 failure", 1)
    assert event.level == "CRITICAL"
    assert event.service == "api"
    assert event.correlation_id == "r-1"


def test_naive_timestamp_uses_utc() -> None:
    parsed = parse_timestamp("2026-01-01 01:02:03")
    assert parsed is not None and parsed.tzinfo == UTC


def test_bad_timestamp_returns_none() -> None:
    assert parse_timestamp("yesterday") is None
