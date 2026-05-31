"use client";
import LinkRowActions from "@/features/dashboard/components/link-row-actions";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { LinkItem } from "../types";
import { cn } from "@/lib/utils";

type LinkTableProps = {
  items: LinkItem[];
  onEdit: (link: LinkItem) => void;
  onStatusChange: (id: number, status: string) => void;
  onDelete: (link: LinkItem) => void;
};

const NEXT_PUBLIC_LINK_BASE_URL =
  process.env.NEXT_PUBLIC_LINK_BASE_URL ?? "https://renice.cc/s/";

export default function LinkTable({
  items,
  onEdit,
  onStatusChange,
  onDelete,
}: LinkTableProps) {
  return (
    <Table className="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead className="w-[20%]">
            <span className="block truncate">原链接</span>
          </TableHead>
          <TableHead className="w-[20%]">
            <span className="block truncate">生成后的短链接</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">浏览量</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">短链接状态</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">创建时间</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">上次更新时间</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">到期时间</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">操作</span>
          </TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        {items.map((item) => (
          <TableRow key={item.id}>
            <TableCell>
              <span className="block truncate">{item.original_url}</span>
            </TableCell>
            <TableCell>
              <span
                className={cn(
                  "block truncate",
                  item.status === "inactive" && "line-through",
                )}
              >
                {NEXT_PUBLIC_LINK_BASE_URL + item.code}
              </span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.view_count}</span>
            </TableCell>
            <TableCell>
              <span
                className={cn(
                  "block truncate",
                  item.status === "inactive" && "text-red-500",
                  item.status === "pending" && "text-gray-400",
                  item.status === "unknown" && "text-gray-400",
                  item.status === "unsafe" && "text-red-500",
                )}
              >
                {item.status === "active"
                  ? "启用"
                  : item.status === "inactive"
                    ? "禁用"
                    : item.status === "pending"
                      ? "审核中"
                      : item.status === "unsafe"
                        ? "不安全"
                        : "未知状态"}
              </span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.created_at}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.updated_at}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">
                {!!item.expires_at ? item.expires_at : "永不过期"}
              </span>
            </TableCell>
            <TableCell>
              <LinkRowActions
                disabled={
                  item.status !== "active" && item.status !== "inactive"
                }
                onEdit={() => onEdit?.(item)}
                onStatusChange={() => onStatusChange?.(item.id, item.status)}
                onDelete={() => onDelete?.(item)}
                itemStatus={item.status}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
