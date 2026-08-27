import { buildApp } from "./app.js";
import { loadConfig } from "./config.js";

const config = loadConfig();
const app = buildApp(config);

const shutdown = async (signal: string): Promise<void> => {
  app.log.info(
    { event: "shutdown.started", signal },
    "Graceful shutdown started",
  );
  await app.close();
  process.exitCode = 0;
};

process.once("SIGINT", () => void shutdown("SIGINT"));
process.once("SIGTERM", () => void shutdown("SIGTERM"));

try {
  await app.listen({ host: config.host, port: config.port });
  if (config.allowRemote)
    app.log.warn(
      { event: "remote_mode.enabled" },
      "Remote mode is enabled; bearer authentication is required",
    );
} catch (error: unknown) {
  app.log.error(
    { event: "startup.failed", errorCode: "STARTUP_FAILED", error },
    "API startup failed",
  );
  process.exitCode = 1;
}
