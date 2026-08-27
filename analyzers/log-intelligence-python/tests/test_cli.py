import json
from io import StringIO
from pathlib import Path

from sba_log_intelligence.cli import run


def test_plugin_protocol(tmp_path: Path) -> None:
    (tmp_path / "app.log").write_text(
        "2026-01-01T00:00:00Z ERROR failed id 1234\n", encoding="utf-8"
    )
    request = (
        json.dumps(
            {
                "protocolVersion": "1",
                "artifact": {"path": "app.log"},
                "context": {"workspaceRoot": str(tmp_path)},
            }
        )
        + "\n"
    )
    output = StringIO()
    assert run(StringIO(request), output, StringIO()) == 0
    assert '"type":"finding"' in output.getvalue()


def test_plugin_rejects_traversal(tmp_path: Path) -> None:
    request = (
        json.dumps(
            {
                "protocolVersion": "1",
                "artifact": {"path": "../secret.log"},
                "context": {"workspaceRoot": str(tmp_path)},
            }
        )
        + "\n"
    )
    output = StringIO()
    assert run(StringIO(request), output, StringIO()) == 0
    assert "ANALYZER_REQUEST_INVALID" in output.getvalue()
