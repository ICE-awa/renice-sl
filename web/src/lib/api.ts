import { getApiErrorMessage } from "./api-error-message";

export type ApiResponse<T> = {
  code: number;
  data?: T;
  message: string;
};

type ApiFetchOptions = RequestInit & {
  skipAuthRefresh?: boolean;
};

export class ApiError<T = unknown> extends Error {
  constructor(
    message: string,
    public readonly code: number,
    public readonly status: number,
    public readonly data?: T,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

let refreshPromise: Promise<unknown> | null = null;

export function refreshAccessToken<T = unknown>(): Promise<T> {
  if (refreshPromise) {
    return refreshPromise as Promise<T>;
  }

  refreshPromise = apiFetch<T>("/api/v1/auth/refresh", {
    method: "POST",
    skipAuthRefresh: true,
  }).finally(() => {
    refreshPromise = null;
  });

  return refreshPromise as Promise<T>;
}

export async function apiFetch<T>(
  path: string,
  init?: ApiFetchOptions,
): Promise<T> {
  const res = await fetch(path, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  const body = (await res.json()) as ApiResponse<T>;

  if (res.status === 401 && !init?.skipAuthRefresh) {
    await apiFetch("/api/v1/auth/refresh", {
      method: "POST",
      skipAuthRefresh: true,
    });

    return apiFetch<T>(path, {
      ...init,
      skipAuthRefresh: true,
    });
  }

  if (!res.ok || body.code !== 0) {
    const message = getApiErrorMessage(body.code, body.message);
    throw new ApiError(message, body.code, res.status, body.data);
  }

  return body.data as T;
}
