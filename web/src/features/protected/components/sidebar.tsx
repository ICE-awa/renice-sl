import { Button } from "@/components/ui/button";
import { SidebarItems } from "../../dashboard/types";
import Link from "next/link";

type SidebarProps = {
  items: SidebarItems[];
};

export default function Sidebar({ items }: SidebarProps) {
  return (
    <div className="flex flex-col gap-2 py-4">
      {items.map((item) => (
        <Button
          key={item.id}
          variant="ghost"
          className="w-full justify-start"
          asChild
        >
          <Link href={item.next}>{item.name}</Link>
        </Button>
      ))}
    </div>
  );
}
