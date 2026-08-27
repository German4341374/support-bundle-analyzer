import { createReadStream } from "node:fs";
import { readFile } from "node:fs/promises";
import { createInterface } from "node:readline";
import { join } from "node:path";
import type { Manifest, Page } from "./types.js";

export async function readManifest(root: string): Promise<Manifest> {
  return JSON.parse(
    await readFile(join(root, "manifest.json"), "utf8"),
  ) as Manifest;
}

export async function readJsonLinesPage(
  filename: string,
  cursor: string | undefined,
  requestedLimit: number | undefined,
): Promise<Page<unknown>> {
  const offset = decodeCursor(cursor);
  const limit = Math.min(Math.max(requestedLimit ?? 50, 1), 200);
  const input = createReadStream(filename, { encoding: "utf8" });
  const lines = createInterface({ input, crlfDelay: Number.POSITIVE_INFINITY });
  const items: unknown[] = [];
  let index = 0;
  let hasMore = false;
  for await (const line of lines) {
    if (index++ < offset) continue;
    if (items.length >= limit) {
      hasMore = true;
      break;
    }
    if (line.trim()) items.push(JSON.parse(line) as unknown);
  }
  lines.close();
  input.destroy();
  return {
    items,
    nextCursor: hasMore ? encodeCursor(offset + items.length) : null,
  };
}

export function encodeCursor(offset: number): string {
  return Buffer.from(String(offset), "utf8").toString("base64url");
}

export function decodeCursor(cursor: string | undefined): number {
  if (!cursor) return 0;
  const value = Number.parseInt(
    Buffer.from(cursor, "base64url").toString("utf8"),
    10,
  );
  if (!Number.isInteger(value) || value < 0) throw new Error("INVALID_CURSOR");
  return value;
}
