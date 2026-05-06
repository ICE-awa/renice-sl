import { apiFetch } from "@/lib/api";
import {
  CreateLinkInput,
  GetLinksInput,
  GetStatsResponse,
  LinkItem,
  UpdateLinkInput,
} from "./types";

export function getLinks(params: GetLinksInput) {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      searchParams.set(key, String(value));
    }
  });

  return apiFetch<LinkItem[]>(`/api/v1/links?${searchParams.toString()}`, {
    method: "GET",
  });
}

export function getLinkByID(id: number) {
  return apiFetch<LinkItem>("/api/v1/link/" + id, {
    method: "GET",
  });
}

export function createLink(params: CreateLinkInput) {
  return apiFetch<void>("/api/v1/link", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(params),
  });
}

export function updateLink(params: UpdateLinkInput) {
  return apiFetch<void>("/api/v1/link/" + params.id, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(params),
  });
}

export function deleteLink(id: number) {
  return apiFetch<void>("/api/v1/link/" + id, {
    method: "DELETE",
  });
}

export function getStats() {
  return apiFetch<GetStatsResponse>("/api/v1/stats", {
    method: "GET",
  });
}
