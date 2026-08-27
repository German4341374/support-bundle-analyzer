from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import TextIO

from .analyzer import analyze_file, findings


def run(stdin: TextIO, stdout: TextIO, stderr: TextIO) -> int:
    for raw in stdin:
        try:
            request = json.loads(raw)
            if not isinstance(request, dict) or request.get("protocolVersion") != "1":
                raise ValueError("unsupported or missing protocolVersion")
            artifact = request.get("artifact")
            if not isinstance(artifact, dict) or not isinstance(artifact.get("path"), str):
                raise ValueError("artifact.path is required")
            context = request.get("context")
            root = Path(".")
            if isinstance(context, dict) and isinstance(context.get("workspaceRoot"), str):
                root = Path(context["workspaceRoot"])
            path = (root / artifact["path"]).resolve()
            if root.resolve() not in path.parents and path != root.resolve():
                raise ValueError("artifact path escapes workspaceRoot")
            result = analyze_file(path)
            for finding in findings(result):
                stdout.write(json.dumps(finding, separators=(",", ":")) + "\n")
            stdout.write(
                json.dumps(
                    {
                        "type": "summary",
                        "lines": result.lines,
                        "malformedLines": result.malformed_lines,
                        "correlationIds": len(result.correlations),
                        "privacy": result.privacy_counts,
                    },
                    separators=(",", ":"),
                )
                + "\n"
            )
            stdout.flush()
        except (OSError, ValueError, json.JSONDecodeError) as error:
            stderr.write(f"plugin request rejected: {error}\n")
            stdout.write(
                json.dumps(
                    {
                        "type": "error",
                        "error": {
                            "code": "ANALYZER_REQUEST_INVALID",
                            "message": str(error),
                        },
                    }
                )
                + "\n"
            )
            stdout.flush()
    return 0


def main() -> None:
    raise SystemExit(run(sys.stdin, sys.stdout, sys.stderr))


if __name__ == "__main__":
    main()
