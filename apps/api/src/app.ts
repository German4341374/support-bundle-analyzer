import { timingSafeEqual } from "node:crypto";
import { mkdir, readFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import Fastify, { type FastifyError, type FastifyInstance } from "fastify";
import type { Config } from "./config.js";
import { execute, runAnalysis } from "./runner.js";
import { SessionStore } from "./store.js";
import { readJsonLinesPage, readManifest } from "./workspace.js";

interface CreateAnalysisBody {
  inputPath: string;
}

interface AnalysisParams {
  id: string;
}

interface PageQuery {
  cursor?: string;
  limit?: number;
}

interface RedactBody {
  profile: "standard" | "strict";
}

interface ComparisonBody {
  baselineAnalysisId: string;
  incidentAnalysisId: string;
}

export function buildApp(
  config: Config,
  store = new SessionStore(),
): FastifyInstance {
  const app = Fastify({
    logger: {
      level: process.env.LOG_LEVEL ?? "info",
      redact: ["req.headers.authorization", "req.headers.cookie"],
    },
    bodyLimit: 64 * 1024,
    requestIdHeader: "x-request-id",
  });

  app.addHook("onRequest", async (request, reply) => {
    if (
      !config.allowRemote ||
      request.url === "/health" ||
      request.url === "/ready"
    )
      return;
    const authorization = request.headers.authorization ?? "";
    const supplied = authorization.startsWith("Bearer ")
      ? authorization.slice(7)
      : "";
    const expected = config.accessToken ?? "";
    const valid =
      supplied.length === expected.length &&
      supplied.length > 0 &&
      timingSafeEqual(Buffer.from(supplied), Buffer.from(expected));
    if (!valid)
      await reply
        .code(401)
        .send(
          errorBody(
            "AUTHENTICATION_REQUIRED",
            "A valid bearer token is required",
          ),
        );
  });

  app.get("/health", () => ({
    status: "ok",
    service: "support-bundle-analyzer-api",
  }));
  app.get("/ready", () => ({ status: "ready" }));
  app.get("/metrics", async (_request, reply) => {
    const sessions = store.list();
    const lines = [
      "# HELP sba_analyses_total Local analysis sessions by status.",
      "# TYPE sba_analyses_total gauge",
      ...["queued", "running", "completed", "failed"].map(
        (status) =>
          `sba_analyses_total{status="${status}"} ${sessions.filter((session) => session.status === status).length}`,
      ),
    ];
    return reply
      .type("text/plain; version=0.0.4")
      .send(`${lines.join("\n")}\n`);
  });

  app.post<{ Body: CreateAnalysisBody }>(
    "/api/v1/analyses",
    {
      schema: {
        body: {
          type: "object",
          additionalProperties: false,
          required: ["inputPath"],
          properties: {
            inputPath: { type: "string", minLength: 1, maxLength: 4096 },
          },
        },
      },
    },
    async (request, reply) => {
      const workspacePath = join(config.workspaceRoot, crypto.randomUUID());
      const session = store.create(
        resolve(request.body.inputPath),
        workspacePath,
      );
      setImmediate(() => void runAnalysis(config, store, session.id));
      return reply.code(202).send(publicSession(session));
    },
  );

  app.get("/api/v1/analyses", () => ({
    items: store.list().map(publicSession),
  }));
  app.get<{ Params: AnalysisParams }>(
    "/api/v1/analyses/:id",
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      let manifest: unknown = null;
      if (session.status === "completed")
        manifest = await readManifest(session.workspacePath);
      return { ...publicSession(session), manifest };
    },
  );

  for (const [route, filename] of [
    ["findings", "findings.jsonl"],
    ["timeline", "timeline.jsonl"],
  ] as const) {
    app.get<{ Params: AnalysisParams; Querystring: PageQuery }>(
      `/api/v1/analyses/:id/${route}`,
      async (request, reply) => {
        const session = store.get(request.params.id);
        if (!session)
          return reply
            .code(404)
            .send(
              errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
            );
        if (session.status !== "completed")
          return reply
            .code(409)
            .send(
              errorBody("ANALYSIS_NOT_COMPLETE", "Analysis has not completed"),
            );
        return readJsonLinesPage(
          join(session.workspacePath, filename),
          request.query.cursor,
          request.query.limit,
        );
      },
    );
  }

  app.get<{ Params: AnalysisParams; Querystring: PageQuery }>(
    "/api/v1/analyses/:id/artifacts",
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      if (session.status !== "completed")
        return reply
          .code(409)
          .send(
            errorBody("ANALYSIS_NOT_COMPLETE", "Analysis has not completed"),
          );
      const manifest = await readManifest(session.workspacePath);
      const offset = request.query.cursor
        ? Number.parseInt(
            Buffer.from(request.query.cursor, "base64url").toString("utf8"),
            10,
          )
        : 0;
      const limit = Math.min(Math.max(request.query.limit ?? 50, 1), 200);
      const items = manifest.artifacts.slice(offset, offset + limit);
      const nextCursor =
        offset + limit < manifest.artifacts.length
          ? Buffer.from(String(offset + limit)).toString("base64url")
          : null;
      return { items, nextCursor };
    },
  );

  app.get<{ Params: AnalysisParams }>(
    "/api/v1/analyses/:id/analyzers",
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      if (session.status !== "completed")
        return reply
          .code(409)
          .send(
            errorBody("ANALYSIS_NOT_COMPLETE", "Analysis has not completed"),
          );
      return { items: (await readManifest(session.workspacePath)).analyzers };
    },
  );

  app.get<{ Params: AnalysisParams }>(
    "/api/v1/analyses/:id/events",
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      reply.hijack();
      reply.raw.writeHead(200, {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache, no-transform",
        Connection: "keep-alive",
      });
      for (const event of session.events) reply.raw.write(formatSSE(event));
      if (session.status === "completed" || session.status === "failed") {
        reply.raw.end();
        return;
      }
      const unsubscribe = store.subscribe(session.id, (event) => {
        reply.raw.write(formatSSE(event));
        if (
          event.stage === "analysis.completed" ||
          event.stage === "analysis.failed"
        )
          reply.raw.end();
      });
      request.raw.once("close", unsubscribe);
    },
  );

  app.post<{ Params: AnalysisParams }>(
    "/api/v1/analyses/:id/report",
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      return {
        reportPath: join(session.workspacePath, "report", "index.html"),
      };
    },
  );

  app.post<{ Params: AnalysisParams; Body: RedactBody }>(
    "/api/v1/analyses/:id/redact",
    {
      schema: {
        body: {
          type: "object",
          additionalProperties: false,
          required: ["profile"],
          properties: { profile: { enum: ["standard", "strict"] } },
        },
      },
    },
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (!session)
        return reply
          .code(404)
          .send(
            errorBody("ANALYSIS_NOT_FOUND", "Analysis session was not found"),
          );
      if (session.status !== "completed")
        return reply
          .code(409)
          .send(
            errorBody("ANALYSIS_NOT_COMPLETE", "Analysis has not completed"),
          );
      const output = join(
        config.workspaceRoot,
        "sanitized",
        `${session.id}-${request.body.profile}`,
      );
      await mkdir(join(config.workspaceRoot, "sanitized"), {
        recursive: true,
        mode: 0o700,
      });
      await execute(
        config.coreBinary,
        [
          "redact",
          session.workspacePath,
          "--profile",
          request.body.profile,
          "--output",
          output,
        ],
        undefined,
        config.analysisTimeoutMs,
      );
      return reply
        .code(201)
        .send({ profile: request.body.profile, path: output });
    },
  );

  app.post<{ Body: ComparisonBody }>(
    "/api/v1/comparisons",
    {
      schema: {
        body: {
          type: "object",
          additionalProperties: false,
          required: ["baselineAnalysisId", "incidentAnalysisId"],
          properties: {
            baselineAnalysisId: { type: "string", minLength: 1 },
            incidentAnalysisId: { type: "string", minLength: 1 },
          },
        },
      },
    },
    async (request, reply) => {
      const baseline = store.get(request.body.baselineAnalysisId);
      const incident = store.get(request.body.incidentAnalysisId);
      if (!baseline || !incident)
        return reply
          .code(404)
          .send(
            errorBody(
              "ANALYSIS_NOT_FOUND",
              "One or more analyses were not found",
            ),
          );
      if (baseline.status !== "completed" || incident.status !== "completed")
        return reply
          .code(409)
          .send(
            errorBody(
              "ANALYSIS_NOT_COMPLETE",
              "Both analyses must be complete",
            ),
          );
      const output = join(
        config.workspaceRoot,
        `comparison-${baseline.id}-${incident.id}.json`,
      );
      await execute(
        config.coreBinary,
        [
          "diff",
          baseline.workspacePath,
          incident.workspacePath,
          "--output",
          output,
        ],
        undefined,
        config.analysisTimeoutMs,
      );
      return reply
        .code(201)
        .send(JSON.parse(await readFile(output, "utf8")) as unknown);
    },
  );

  app.get("/api/v1/plugins", () => ({ items: [], protocolVersion: "1" }));
  app.get("/api/v1/rules", () => ({
    items: [
      "LOG_CONNECTION_REFUSED",
      "LOG_DATABASE_ERROR",
      "LOG_AUTHENTICATION_FAILED",
      "LOG_TIMEOUT",
      "LOG_OUT_OF_MEMORY",
      "LOG_DNS_FAILURE",
    ],
  }));

  app.setErrorHandler((error: FastifyError, _request, reply) => {
    const status =
      "statusCode" in error && typeof error.statusCode === "number"
        ? error.statusCode
        : 500;
    void reply
      .code(status)
      .send(
        errorBody(
          status >= 500 ? "INTERNAL_ERROR" : "INVALID_REQUEST",
          status >= 500 ? "The request could not be completed" : error.message,
        ),
      );
  });
  return app;
}

function publicSession(session: ReturnType<SessionStore["create"]>) {
  return {
    id: session.id,
    status: session.status,
    createdAt: session.createdAt,
    completedAt: session.completedAt,
    errorCode: session.errorCode,
  };
}

function errorBody(code: string, message: string) {
  return { error: { code, message } };
}

function formatSSE(event: object): string {
  return `event: progress\ndata: ${JSON.stringify(event)}\n\n`;
}
