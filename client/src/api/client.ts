import type {
  AstroObject,
  ApiErrorBody,
  CreateMissionResponse,
  Mission,
  MissionData,
  MissionProgressEvent,
  UnauthorizedErrorBody,
} from "@/types/api";

export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

export interface AstrohopClientOptions {
  baseUrl: string;
  token?: string;
  fetchFn?: typeof fetch;
}

export class AstrohopClient {
  private baseUrl: string;
  private token: string | null;
  private fetchFn: typeof fetch;

  constructor(options: AstrohopClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.token = options.token ?? null;
    this.fetchFn = options.fetchFn ?? fetch.bind(globalThis);
  }

  setToken(token: string | null): void {
    this.token = token;
  }

  getToken(): string | null {
    return this.token;
  }

  private authHeaders(): Record<string, string> {
    return this.token ? { Authorization: `Bearer ${this.token}` } : {};
  }

  private async request<T>(
    path: string,
    init: RequestInit & { expectBody?: boolean } = {},
  ): Promise<T> {
    const { expectBody = true, ...requestInit } = init;
    const response = await this.fetchFn(`${this.baseUrl}${path}`, {
      ...requestInit,
      headers: {
        ...(requestInit.body ? { "Content-Type": "application/json" } : {}),
        ...this.authHeaders(),
        ...(requestInit.headers ?? {}),
      },
    });

    if (!response.ok) {
      let body: unknown = undefined;
      try {
        body = await response.json();
      } catch {}
      const message =
        (body as ApiErrorBody | UnauthorizedErrorBody | undefined) &&
        ("error" in (body as any)
          ? (body as ApiErrorBody).error
          : (body as any).message);
      throw new ApiError(
        response.status,
        message ?? `Request to ${path} failed with status ${response.status}`,
        body,
      );
    }

    if (!expectBody || response.status === 204) {
      return undefined as T;
    }
    return (await response.json()) as T;
  }

  async createAccount(): Promise<{ token: string }> {
    const response = await this.fetchFn(`${this.baseUrl}/api/v1/accounts`, {
      method: "POST",
    });

    if (!response.ok) {
      throw new ApiError(
        response.status,
        `Account creation failed with status ${response.status}`,
        undefined,
      );
    }

    const authHeader = response.headers.get("Authorization");
    if (!authHeader) {
      throw new ApiError(
        response.status,
        "Account created but no Authorization header was returned",
        undefined,
      );
    }

    const token = authHeader.replace(/^Bearer\s+/i, "");
    this.token = token;
    return { token };
  }

  async searchObjects(name: string): Promise<AstroObject[]> {
    const query = new URLSearchParams({ name });
    return this.request<AstroObject[]>(
      `/api/v1/search/objects?${query.toString()}`,
      {
        method: "GET",
      },
    );
  }

  async listMissions(): Promise<Mission[]> {
    return this.request<Mission[]>("/api/v1/missions", { method: "GET" });
  }

  async createMission(data: MissionData): Promise<CreateMissionResponse> {
    return this.request<CreateMissionResponse>("/api/v1/missions", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getMission(missionId: string): Promise<Mission> {
    return this.request<Mission>(
      `/api/v1/missions/${encodeURIComponent(missionId)}`,
      {
        method: "GET",
      },
    );
  }

  async updateMission(missionId: string, data: MissionData): Promise<void> {
    await this.request<void>(
      `/api/v1/missions/${encodeURIComponent(missionId)}`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
        expectBody: false,
      },
    );
  }

  async deleteMission(missionId: string): Promise<void> {
    await this.request<void>(
      `/api/v1/missions/${encodeURIComponent(missionId)}`,
      {
        method: "DELETE",
        expectBody: false,
      },
    );
  }

  async updateMissionVisibility(
    missionId: string,
    isPublic: boolean,
  ): Promise<void> {
    const query = new URLSearchParams({ public: String(isPublic) });
    await this.request<void>(
      `/api/v1/missions/${encodeURIComponent(missionId)}/visibility?${query.toString()}`,
      { method: "PATCH", expectBody: false },
    );
  }

  streamMissionProgress(
    missionId: string,
    handlers: {
      onEvent: (event: MissionProgressEvent) => void;
      onError?: (error: unknown) => void;
      onClose?: () => void;
    },
  ): AbortController {
    const controller = new AbortController();

    (async () => {
      try {
        const response = await this.fetchFn(
          `${this.baseUrl}/api/v1/missions/${encodeURIComponent(missionId)}/stream`,
          {
            method: "GET",
            headers: { ...this.authHeaders(), Accept: "text/event-stream" },
            signal: controller.signal,
          },
        );

        if (!response.ok || !response.body) {
          let body: unknown;
          try {
            body = await response.json();
          } catch {}
          throw new ApiError(
            response.status,
            `Stream request failed with status ${response.status}`,
            body,
          );
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });

          let boundary: number;
          while ((boundary = buffer.indexOf("\n\n")) !== -1) {
            const rawFrame = buffer.slice(0, boundary);
            buffer = buffer.slice(boundary + 2);

            const dataLines = rawFrame
              .split("\n")
              .filter((line) => line.startsWith("data:"))
              .map((line) => line.slice(5).trimStart());

            if (dataLines.length === 0) continue;

            const dataStr = dataLines.join("\n");
            try {
              const parsed = JSON.parse(dataStr) as MissionProgressEvent;
              handlers.onEvent(parsed);
            } catch (parseErr) {
              handlers.onError?.(parseErr);
            }
          }
        }

        handlers.onClose?.();
      } catch (err) {
        if ((err as { name?: string }).name === "AbortError") {
          handlers.onClose?.();
          return;
        }
        handlers.onError?.(err);
      }
    })();

    return controller;
  }
}

const apiAddr = import.meta.env.VITE_API_ADDR;

if (!apiAddr) {
  throw new Error("Missing API_ADDR: set VITE_API_ADDR");
}

export const apiClient = new AstrohopClient({ baseUrl: apiAddr });
