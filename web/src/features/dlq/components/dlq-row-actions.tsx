import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

type DLQRowActionsProps = {
  onRetry: () => void;
  onResolve: () => void;
};

export default function DLQRowActions({
  onRetry,
  onResolve,
}: DLQRowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="sm" variant="default">
          ...
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuItem onClick={onRetry}>重试</DropdownMenuItem>
        <DropdownMenuItem onClick={onResolve}>标记为已解决</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
