import { apiFetch } from "@/lib/api";
import { GetDLQMessagesResponse, GetDLQMessagesParams } from "./types";

export function getDLQMessages(params: GetDLQMessagesParams) {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) {
      searchParams.set(key, String(value));
    }
  });

  return apiFetch<GetDLQMessagesResponse>(
    "/api/v1/admin/dlq?" + searchParams.toString(),
    {
      method: "GET",
    },
  );
}

export function retryDLQMessage(id: number) {
  return apiFetch<void>("/api/v1/admin/dlq/retry/" + id, {
    method: "POST",
  });
}

export function resolveDLQMessage(id: number) {
  return apiFetch<void>("/api/v1/admin/dlq/resolve/" + id, {
    method: "POST",
  });
}
