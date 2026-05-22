"use client";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useSession } from "./session-provider";

export type SidebarItems = {
  id: number;
  name: string;
  next: string;
  roles: string[];
};

export default function Sidebar() {
  const pathname = usePathname();

  const items = [
    {
      id: 1,
      name: "链接管理",
      next: "/dashboard",
      roles: ["user", "admin"],
    },
    {
      id: 2,
      name: "状态图表",
      next: "/charts",
      roles: ["admin"],
    },
  ];

  const { user } = useSession();

  return (
    <div className="flex flex-col gap-2 py-4">
      {items
        .filter((item) => item.roles.includes(user.role))
        .map((item) => {
          const active =
            pathname === item.next || pathname.startsWith(item.next + "/");

          return (
            <Button
              key={item.id}
              variant={active ? "secondary" : "ghost"}
              className="w-full justify-start"
              asChild
            >
              <Link href={item.next}>{item.name}</Link>
            </Button>
          );
        })}
    </div>
  );
}
