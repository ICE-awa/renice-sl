export type ApiResponse<T> = {
  code: number;
  data?: T;
  message: string;
};

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
    throw new Error(body.message);
  }

  return body.data as T;
}
