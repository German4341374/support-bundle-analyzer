import { mkdtemp, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { readJsonLinesPage } from "../src/workspace.js";

describe("cursor pagination", () => {
  it("returns stable bounded pages", async () => {
    const directory = await mkdtemp(join(tmpdir(), "sba-api-test-"));
    const filename = join(directory, "items.jsonl");
    await writeFile(filename, '{"id":1}\n{"id":2}\n{"id":3}\n', {
      mode: 0o600,
    });
    const first = await readJsonLinesPage(filename, undefined, 2);
    expect(first.items).toEqual([{ id: 1 }, { id: 2 }]);
    expect(first.nextCursor).not.toBeNull();
    const second = await readJsonLinesPage(
      filename,
      first.nextCursor ?? undefined,
      2,
    );
    expect(second.items).toEqual([{ id: 3 }]);
  });
});
