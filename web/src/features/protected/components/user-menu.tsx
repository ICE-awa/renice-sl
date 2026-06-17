"use client";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { logout } from "@/features/auth/api";
import { clearScheduledRefresh } from "@/features/protected/components/session";
import { ApiError } from "@/lib/api";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useSession } from "./session-provider";

function getGreeting(hour: number) {
  if (hour >= 5 && hour < 11) return "早上好";
  else if (hour >= 11 && hour < 14) return "中午好";
  else if (hour >= 14 && hour < 18) return "下午好";
  else if (hour >= 18 && hour < 22) return "晚上好";
  else return "夜深了";
}

export function UserMenu() {
  const hour = new Date().getHours();

  const { user } = useSession();

  const router = useRouter();

  async function onLogout() {
    try {
      clearScheduledRefresh();
      await logout();
      toast.success("退出登录成功！");
      router.push("/login");
    } catch (err) {
      const message = err instanceof ApiError ? err.message : "服务器打了个盹";
      toast.error(message);
    }
  }

  return (
    <div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="w-full justify-between">
            <Avatar>
              <AvatarFallback>
                {user.username !== undefined ? user.username.charAt(0) : "I"}
              </AvatarFallback>
            </Avatar>
            <span>
              {getGreeting(hour)} {user.username}
            </span>
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent>
          <DropdownMenuItem onClick={onLogout}>
            <span className="text-destructive!">退出登录</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
