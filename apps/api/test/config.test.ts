import { describe, expect, it } from "vitest";
import { loadConfig } from "../src/config.js";

describe("configuration", () => {
  it("binds to loopback by default", () => {
    expect(loadConfig({}).host).toBe("127.0.0.1");
  });

  it("rejects remote binding without explicit opt-in", () => {
    expect(() => loadConfig({ SBA_HOST: "0.0.0.0" })).toThrow(
      "SBA_ALLOW_REMOTE",
    );
  });

  it("requires a sufficiently long token for remote mode", () => {
    expect(() =>
      loadConfig({
        SBA_HOST: "0.0.0.0",
        SBA_ALLOW_REMOTE: "true",
        SBA_ACCESS_TOKEN: "short",
      }),
    ).toThrow("SBA_ACCESS_TOKEN");
  });
});
