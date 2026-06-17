import { apiFetch } from "@/lib/api";
import { StatRequest, StatResponse } from "./type";

export function getClickStat(params: StatRequest) {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) {
      searchParams.set(key, value.toString());
    }
  });

  return apiFetch<StatResponse[]>(
    `/api/v1/admin/stats/click?${searchParams.toString()}`,
    {
      method: "GET",
    },
  );
}

export function getUserStat(params: StatRequest) {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) {
      searchParams.set(key, value.toString());
    }
  });

  return apiFetch<StatResponse[]>(
    `/api/v1/admin/stats/user?${searchParams.toString()}`,
    {
      method: "GET",
    },
  );
}

export function getLinkStat(params: StatRequest) {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) {
      searchParams.set(key, value.toString());
    }
  });

  return apiFetch<StatResponse[]>(
    `/api/v1/admin/stats/link?${searchParams.toString()}`,
    {
      method: "GET",
    },
  );
}
