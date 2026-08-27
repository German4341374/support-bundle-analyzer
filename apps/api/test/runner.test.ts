import { describe, expect, it } from "vitest";
import { execute } from "../src/runner.js";

describe("analyzer process runner", () => {
  it("terminates a process that exceeds its deadline", async () => {
    await expect(
      execute(
        process.execPath,
        ["-e", "setTimeout(() => {}, 10000)"],
        undefined,
        25,
      ),
    ).rejects.toThrow("exceeded timeout");
  });
});
