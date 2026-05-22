"use client";
import type { CurrentUser } from "@/features/auth/types";
import { me } from "@/features/auth/api";
import { useRouter } from "next/navigation";
import {
  createContext,
  useContext,
  useMemo,
  useEffect,
  useState,
  type ReactNode,
} from "react";

type SessionContextValue = {
  user: CurrentUser;
  setUser: (user: CurrentUser) => void;
  isAdmin: boolean;
};

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<CurrentUser>({} as CurrentUser);

  const value = useMemo(
    () => ({
      user,
      setUser,
      isAdmin: user.role === "admin",
    }),
    [user],
  );

  useEffect(() => {
    let alive = true;

    async function fetchCurrentUser() {
      await me()
        .then((resp) => {
          if (!alive) return;
          setUser(resp);
        })
        .catch(() => {
          if (!alive) return;
          router.replace("/login");
        });
    }
    void fetchCurrentUser();

    return () => {
      alive = false;
    };
  }, [router]);

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  );
}

export function useSession() {
  const ctx = useContext(SessionContext);

  if (!ctx) {
    throw new Error("useSession must be used within a SessionProvider");
  }

  return ctx;
}
