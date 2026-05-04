import { refresh } from "../../auth/api";

let refreshTimer: ReturnType<typeof setTimeout> | undefined;

export function scheduleRefresh(expiresIn: number) {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
  }

  const delay = Math.max((expiresIn - 60) * 1000, 0);

  refreshTimer = setTimeout(async () => {
    try {
      const resp = await refresh();
      scheduleRefresh(resp.expires_in);
    } catch {
      clearScheduledRefresh();
    }
  }, delay);
}

export function clearScheduledRefresh() {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
    refreshTimer = undefined;
  }
}
