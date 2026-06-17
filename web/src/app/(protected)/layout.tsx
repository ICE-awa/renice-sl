import SessionRefresher from "@/features/protected/components/session-refresher";
import { SessionProvider } from "@/features/protected/components/session-provider";
import Sidebar from "@/features/protected/components/sidebar";
import { UserMenu } from "@/features/protected/components/user-menu";
import { ReactNode } from "react";

export default async function Layout({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <SessionRefresher>
        <div className="flex h-svh">
          <aside className="flex min-h-svh w-72 flex-col border-r bg-muted/30 px-3 py-8">
            <div className="flex h-14 items-center justify-center border-b">
              <span className="text-lg font-semibold">renice 短链接</span>
            </div>
            <nav className="mt-4 flex flex-col flex-1 gap-1">
              <Sidebar />
            </nav>
            <UserMenu />
          </aside>
          <main className="flex min-h-0 flex-1 flex-col p-6">{children}</main>
        </div>
      </SessionRefresher>
    </SessionProvider>
  );
}
