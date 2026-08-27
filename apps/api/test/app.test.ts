import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { buildApp } from "../src/app.js";
import type { Config } from "../src/config.js";

const config: Config = {
  host: "127.0.0.1",
  port: 8080,
  allowRemote: false,
  workspaceRoot: resolve(".tmp/api-tests"),
  inputRoot: resolve("."),
  coreBinary: resolve(".tmp/missing-binary"),
  analysisTimeoutMs: 5_000,
};

describe("API", () => {
  it("reports health", async () => {
    const app = buildApp(config);
    const response = await app.inject({ method: "GET", url: "/health" });
    expect(response.statusCode).toBe(200);
    expect(response.json()).toMatchObject({ status: "ok" });
    await app.close();
  });

  it("returns normalized not-found errors", async () => {
    const app = buildApp(config);
    const response = await app.inject({
      method: "GET",
      url: "/api/v1/analyses/missing",
    });
    expect(response.statusCode).toBe(404);
    expect(response.json()).toEqual({
      error: {
        code: "ANALYSIS_NOT_FOUND",
        message: "Analysis session was not found",
      },
    });
    await app.close();
  });

  it("validates analysis bodies", async () => {
    const app = buildApp(config);
    const response = await app.inject({
      method: "POST",
      url: "/api/v1/analyses",
      payload: {},
    });
    expect(response.statusCode).toBe(400);
    expect(response.json()).toHaveProperty("error.code", "INVALID_REQUEST");
    await app.close();
  });

  it("protects remote mode with constant-time bearer comparison", async () => {
    const remote = buildApp({
      ...config,
      allowRemote: true,
      accessToken: "012345678901234567890123",
    });
    expect(
      (await remote.inject({ method: "GET", url: "/ready" })).statusCode,
    ).toBe(200);
    expect(
      (
        await remote.inject({
          method: "GET",
          url: "/api/v1/analyses",
        })
      ).statusCode,
    ).toBe(401);
    expect(
      (
        await remote.inject({
          method: "GET",
          url: "/api/v1/analyses",
          headers: { authorization: "Bearer 012345678901234567890123" },
        })
      ).statusCode,
    ).toBe(200);
    await remote.close();
  });
});
