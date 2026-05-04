import ProtectedSession from "@/features/protected/components/protected-session";
import Sidebar from "@/features/protected/components/sidebar";
import { UserMenu } from "@/features/protected/components/user-menu";
import { ReactNode } from "react";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const sidebarItems = [
    {
      id: 1,
      name: "导航栏 1",
      next: "/test1",
    },
    {
      id: 2,
      name: "导航栏 2",
      next: "/test2",
    },
  ];

  const user = "ice";

  return (
    <ProtectedSession>
      <div className="flex min-h-svh">
        <aside className="flex min-h-svh w-72 flex-col border-r bg-muted/30 px-3 py-8">
          <div className="flex h-14 items-center justify-center border-b">
            <span className="text-lg font-semibold">renice 短链接</span>
          </div>
          <nav className="mt-4 flex flex-col flex-1 gap-1">
            <Sidebar items={sidebarItems} />
          </nav>
          <UserMenu user={user} />
        </aside>
        <main className="flex-1 p-6">{children}</main>
      </div>
    </ProtectedSession>
  );
}
