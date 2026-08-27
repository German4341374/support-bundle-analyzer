import { spawn } from "node:child_process";
import { mkdir, realpath } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";
import type { Config } from "./config.js";
import type { SessionStore } from "./store.js";

export function assertWithinRoot(candidate: string, root: string): string {
  const resolved = resolve(candidate);
  const relation = relative(resolve(root), resolved);
  if (
    relation === ".." ||
    relation.startsWith(`..${sep}`) ||
    relation.startsWith(sep)
  ) {
    throw new Error("INPUT_OUTSIDE_ALLOWED_ROOT");
  }
  return resolved;
}

export async function runAnalysis(
  config: Config,
  store: SessionStore,
  id: string,
): Promise<void> {
  const session = store.require(id);
  try {
    const input = assertWithinRoot(
      await realpath(session.inputPath),
      config.inputRoot,
    );
    await mkdir(dirname(session.workspacePath), {
      recursive: true,
      mode: 0o700,
    });
    store.update(id, { status: "running" });
    store.event(id, "analysis.started", "Analysis process started");
    await execute(
      config.coreBinary,
      ["analyze", "--output", session.workspacePath, input],
      (line) => {
        const match = /^\[([^\]]+)]\s+(.+)$/.exec(line);
        store.event(
          id,
          match?.[1] ?? "analyzer.output",
          match?.[2] ?? "Analyzer reported progress",
        );
      },
    );
    store.update(id, {
      status: "completed",
      completedAt: new Date().toISOString(),
    });
    store.event(id, "analysis.completed", "Analysis completed");
  } catch (error: unknown) {
    const message =
      error instanceof Error ? error.message : "Unknown analyzer failure";
    store.update(id, {
      status: "failed",
      completedAt: new Date().toISOString(),
      errorCode: "ANALYZER_FAILED",
    });
    store.event(id, "analysis.failed", message.slice(0, 300));
  }
}

export function execute(
  binary: string,
  args: string[],
  progress?: (line: string) => void,
): Promise<void> {
  return new Promise((resolvePromise, reject) => {
    const child = spawn(binary, args, {
      shell: false,
      stdio: ["ignore", "ignore", "pipe"],
      windowsHide: true,
    });
    let stderr = "";
    child.stderr.setEncoding("utf8");
    child.stderr.on("data", (chunk: string) => {
      stderr = (stderr + chunk).slice(-16_384);
      for (const line of chunk.split(/\r?\n/).filter(Boolean)) progress?.(line);
    });
    child.once("error", reject);
    child.once("exit", (code, signal) => {
      if (code === 0) resolvePromise();
      else
        reject(
          new Error(
            `analyzer exited with code ${String(code)} signal ${String(signal)}: ${stderr}`,
          ),
        );
    });
  });
}
