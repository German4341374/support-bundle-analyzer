export type AnalysisStatus = "queued" | "running" | "completed" | "failed";

export interface ProgressEvent {
  sequence: number;
  stage: string;
  message: string;
  timestamp: string;
}

export interface AnalysisSession {
  id: string;
  inputPath: string;
  workspacePath: string;
  status: AnalysisStatus;
  createdAt: string;
  completedAt?: string;
  errorCode?: string;
  events: ProgressEvent[];
}

export interface Manifest {
  analysisId: string;
  artifactCount: number;
  inputName: string;
  started: string;
  completed: string;
  warnings: string[];
  artifacts: unknown[];
  analyzers: unknown[];
}

export interface Page<T> {
  items: T[];
  nextCursor: string | null;
}
