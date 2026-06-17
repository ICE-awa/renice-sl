import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

type LinkRowActionsProps = {
  disabled?: boolean;
  onEdit: () => void;
  onStatusChange: () => void;
  onDelete: () => void;
  itemStatus: string;
};

export default function LinkRowActions({
  disabled,
  onEdit,
  onStatusChange,
  onDelete,
  itemStatus,
}: LinkRowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="sm" variant="default" disabled={disabled}>
          ...
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuItem onClick={onEdit}>编辑</DropdownMenuItem>
        <DropdownMenuItem onClick={onStatusChange}>
          {itemStatus === "active" ? "禁用" : "启用"}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onDelete}>
          <span className="text-destructive!">删除</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
