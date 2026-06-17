"use client";

import { refresh } from "@/features/auth/api";
import { useRouter } from "next/navigation";
import { ReactNode, useEffect } from "react";
import { clearScheduledRefresh, scheduleRefresh } from "./session";

export default function SessionRefresher({
  children,
}: {
  children: ReactNode;
}) {
  const router = useRouter();

  useEffect(() => {
    let alive = true;

    refresh()
      .then((resp) => {
        if (!alive) return;
        scheduleRefresh(resp.expires_in);
      })
      .catch(() => {
        clearScheduledRefresh();
        router.replace("/login");
      });

    return () => {
      alive = false;
      clearScheduledRefresh();
    };
  }, [router]);

  return <>{children}</>;
}
