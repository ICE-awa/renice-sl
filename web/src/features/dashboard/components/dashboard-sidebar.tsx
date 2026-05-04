"use client";

import { Button } from "@/components/ui/button";
import { SidebarItems } from "../types";
import Link from "next/link";

type DashboardSidebarProps = {
  items: SidebarItems[];
};

export default function DashboardSidebar({ items }: DashboardSidebarProps) {
  return (
    <div className="flex flex-col gap-2 py-4">
      {items.map((item) => (
        <Button key={item.id} variant="ghost" className="w-full justify-start">
          <Link href={item.next}>{item.name}</Link>
        </Button>
      ))}
    </div>
  );
}
