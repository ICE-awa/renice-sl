import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import DLQRowActions from "./dlq-row-actions";
import type { DLQMessageItem } from "../types";

type MessageTableProps = {
  items: DLQMessageItem[];
  onRetry: (item: DLQMessageItem) => void;
  onResolve: (item: DLQMessageItem) => void;
};

function formatTime(value: string | null) {
  if (value === null) {
    return "N/A";
  }
  const date = new Date(value);
  return new Intl.DateTimeFormat("zh-CN", {
    timeZone: "Asia/Shanghai",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).format(date);
}

export default function MessageTable({
  items,
  onRetry,
  onResolve,
}: MessageTableProps) {
  return (
    <Table className="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead className="w-[10%]">
            <span className="block truncate">Source Stream</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">Source Consumer</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">Stream Seq</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">Subject</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">Payload</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">失败原因</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">状态</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">失败时间</span>
          </TableHead>
          <TableHead className="w-[10%]">
            <span className="block truncate">解决时间</span>
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
              <span className="block truncate">{item.source_stream}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.source_consumer}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.stream_seq}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.subject}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">
                {JSON.stringify(item.payload)}
              </span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.reason}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">{item.status}</span>
            </TableCell>
            <TableCell>
              <span className="block truncate">
                {formatTime(item.failed_at)}
              </span>
            </TableCell>
            <TableCell>
              <span className="block truncate">
                {formatTime(item.resolved_at)}
              </span>
            </TableCell>
            <TableCell>
              <DLQRowActions
                onRetry={() => onRetry(item)}
                onResolve={() => onResolve(item)}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
