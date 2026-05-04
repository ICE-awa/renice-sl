"use client";

import { refresh } from "@/features/auth/api";
import { useRouter } from "next/navigation";
import { ReactNode, useEffect, useState } from "react";
import { clearScheduledRefresh, scheduleRefresh } from "./session";

export default function ProtectedSession({
  children,
}: {
  children: ReactNode;
}) {
  const router = useRouter();

  const [ready, setReady] = useState(false);

  useEffect(() => {
    let alive = true;

    refresh()
      .then((resp) => {
        if (!alive) return;
        scheduleRefresh(resp.expires_in);
        setReady(true);
      })
      .catch((err) => {
        console.warn("[auth] 初始化 refresh 计划失败", err);
        clearScheduledRefresh();
        router.replace("/login");
      });

    return () => {
      alive = false;
      clearScheduledRefresh();
    };
  }, [router]);

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center">
        loading...
      </div>
    );
  }

  return <>{children}</>;
}
