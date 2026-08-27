from __future__ import annotations

import gzip
from pathlib import Path

from sba_log_intelligence.analyzer import analyze_file, findings, normalize_message


def test_normalization_removes_volatile_values() -> None:
    first = normalize_message("2026-01-01T00:00:00Z request_id=a-1 error user 1234 from 192.0.2.1")
    second = normalize_message(
        "2026-01-01T00:00:01Z request_id=b-2 error user 9876 from 198.51.100.1"
    )
    assert first == second


def test_streaming_analysis_groups_errors(tmp_path: Path) -> None:
    path = tmp_path / "app.log"
    path.write_text(
        "2026-01-01T00:00:00Z ERROR service=api request_id=a failed id 1234\n"
        "2026-01-01T00:00:01Z ERROR service=api request_id=b failed id 5678\n",
        encoding="utf-8",
    )
    result = analyze_file(path)
    assert result.lines == 2
    assert list(result.groups.values())[0].count == 2
    assert len(findings(result)) == 1


def test_gzip_input(tmp_path: Path) -> None:
    path = tmp_path / "app.log.gz"
    with gzip.open(path, "wt", encoding="utf-8") as stream:
        stream.write('{"timestamp":"2026-01-01T00:00:00Z","level":"ERROR","message":"failed"}\n')
    assert analyze_file(path).lines == 1


def test_burst_detection(tmp_path: Path) -> None:
    path = tmp_path / "burst.log"
    path.write_text(
        "".join(f"2026-01-01T00:00:0{index}Z ERROR failure {index}\n" for index in range(5)),
        encoding="utf-8",
    )
    assert len(analyze_file(path).bursts) == 1


def test_privacy_detection_does_not_emit_values(tmp_path: Path) -> None:
    path = tmp_path / "privacy.log"
    path.write_text("INFO contact=user@example.test password=fictional-only\n", encoding="utf-8")
    result = analyze_file(path)
    assert result.privacy_counts == {"email": 1, "password": 1}
