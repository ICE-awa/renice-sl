export type ApiResponse<T> = {
  code: number;
  data?: T;
  message: string;
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

export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
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

  if (!res.ok || body.code !== 0) {
    throw new ApiError(body.message, body.code, res.status, body.data);
  }

  return body.data as T;
}
