import { isIP } from "node:net";
import { resolve } from "node:path";

export interface Config {
  host: string;
  port: number;
  allowRemote: boolean;
  accessToken?: string;
  workspaceRoot: string;
  inputRoot: string;
  coreBinary: string;
}

const loopbackHosts = new Set(["127.0.0.1", "::1", "localhost"]);

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const host = env.SBA_HOST ?? "127.0.0.1";
  const port = Number.parseInt(env.SBA_PORT ?? "8080", 10);
  const allowRemote = env.SBA_ALLOW_REMOTE === "true";
  const accessToken = env.SBA_ACCESS_TOKEN;
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error("SBA_PORT must be an integer between 1 and 65535");
  }
  if (!loopbackHosts.has(host) && !allowRemote) {
    throw new Error("A non-loopback SBA_HOST requires SBA_ALLOW_REMOTE=true");
  }
  if (allowRemote && (!accessToken || accessToken.length < 24)) {
    throw new Error(
      "Remote mode requires SBA_ACCESS_TOKEN with at least 24 characters",
    );
  }
  if (isIP(host) === 0 && host !== "localhost") {
    throw new Error("SBA_HOST must be localhost or an IP address");
  }
  return {
    host,
    port,
    allowRemote,
    ...(accessToken ? { accessToken } : {}),
    workspaceRoot: resolve(env.SBA_WORKSPACE_ROOT ?? ".sba/workspaces"),
    inputRoot: resolve(env.SBA_INPUT_ROOT ?? "."),
    coreBinary: resolve(env.SBA_CORE_BINARY ?? "support-bundle-analyzer"),
  };
}
