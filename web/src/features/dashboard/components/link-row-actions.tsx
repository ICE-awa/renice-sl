import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

type LinkRowActionsProps = {
  onEdit?: () => void;
  onDisable?: () => void;
  onDelete?: () => void;
};

export default function LinkRowActions({
  onEdit,
  onDisable,
  onDelete,
}: LinkRowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="sm" variant="default">
          ...
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuItem onClick={onEdit}>编辑</DropdownMenuItem>
        <DropdownMenuItem onClick={onDisable}>禁用</DropdownMenuItem>
        <DropdownMenuItem onClick={onDelete}>
          <span className="text-destructive!">删除</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
