import { randomUUID } from "node:crypto";
import type { AnalysisSession, ProgressEvent } from "./types.js";

export class SessionStore {
  readonly #sessions = new Map<string, AnalysisSession>();
  readonly #listeners = new Map<string, Set<(event: ProgressEvent) => void>>();

  create(inputPath: string, workspacePath: string): AnalysisSession {
    const session: AnalysisSession = {
      id: randomUUID(),
      inputPath,
      workspacePath,
      status: "queued",
      createdAt: new Date().toISOString(),
      events: [],
    };
    this.#sessions.set(session.id, session);
    return session;
  }

  list(): AnalysisSession[] {
    return [...this.#sessions.values()].sort((a, b) =>
      b.createdAt.localeCompare(a.createdAt),
    );
  }

  get(id: string): AnalysisSession | undefined {
    return this.#sessions.get(id);
  }

  update(
    id: string,
    patch: Partial<Omit<AnalysisSession, "id" | "events">>,
  ): void {
    const session = this.require(id);
    Object.assign(session, patch);
  }

  event(id: string, stage: string, message: string): ProgressEvent {
    const session = this.require(id);
    const event: ProgressEvent = {
      sequence: session.events.length + 1,
      stage,
      message,
      timestamp: new Date().toISOString(),
    };
    session.events.push(event);
    for (const listener of this.#listeners.get(id) ?? []) listener(event);
    return event;
  }

  subscribe(id: string, listener: (event: ProgressEvent) => void): () => void {
    this.require(id);
    const listeners = this.#listeners.get(id) ?? new Set();
    listeners.add(listener);
    this.#listeners.set(id, listeners);
    return () => listeners.delete(listener);
  }

  require(id: string): AnalysisSession {
    const session = this.#sessions.get(id);
    if (!session) throw new Error("ANALYSIS_NOT_FOUND");
    return session;
  }
}
