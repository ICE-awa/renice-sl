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

type LinkTableProps = {
  items: LinkItem[];
  onEdit?: (link: LinkItem) => void;
  onDisable?: (link: LinkItem) => void;
  onDelete?: (link: LinkItem) => void;
};

export default function LinkTable({
  items,
  onEdit,
  onDisable,
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
          <TableHead className="w-[13%]">
            <span className="block truncate">创建时间</span>
          </TableHead>
          <TableHead className="w-[13%]">
            <span className="block truncate">上次更新时间</span>
          </TableHead>
          <TableHead className="w-[13%]">
            <span className="block truncate">到期时间</span>
          </TableHead>
          <TableHead className="w-[11%]">
            <span className="block truncate">操作</span>
          </TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        {items.map((item) => (
          <TableRow key={item.id}>
            <TableCell>
              <span className="block truncate">{item.originalUrl}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.shortUrl}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.viewCount}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.createdAt}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.updatedAt}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.expiresAt}</span>
            </TableCell>
            <TableCell>
              <LinkRowActions
                onEdit={() => onEdit?.(item)}
                onDisable={() => onDisable?.(item)}
                onDelete={() => onDelete?.(item)}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
