import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

type LinkRowActionsProps = {
  onEdit: () => void;
  onStatusChange: () => void;
  onDelete: () => void;
  itemStatus: "active" | "inactive";
};

export default function LinkRowActions({
  onEdit,
  onStatusChange,
  onDelete,
  itemStatus,
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
