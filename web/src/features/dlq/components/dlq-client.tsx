"use client";
import Pagination from "@/components/pagination";
import MessageTable from "./message-table";
import { useEffect, useState } from "react";
import { getDLQMessages, retryDLQMessage, resolveDLQMessage } from "../api";
import { ApiError } from "@/lib/api";
import { toast } from "sonner";
import { useSession } from "@/features/protected/components/session-provider";
import { useRouter } from "next/navigation";
import type {
  GetDLQMessagesParams,
  GetDLQMessagesResponse,
  DLQMessageItem,
} from "../types";

export default function DLQClient() {
  const [data, setData] = useState<GetDLQMessagesResponse>({
    total: 0,
    items: [],
    page_num: 1,
    page_size: 10,
  });
  const [searchParams, setSearchParams] = useState<GetDLQMessagesParams>({
    page_num: 1,
    page_size: 10,
  });

  const { user } = useSession();
  const router = useRouter();

  async function onRefreshTable(params: GetDLQMessagesParams) {
    const data = await getDLQMessages(params);
    setData(data);
  }

  async function handleRetry(item: DLQMessageItem) {
    try {
      await retryDLQMessage(item.id);
      await onRefreshTable(searchParams);
      toast.success("正在重试该消息");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function handleResolve(item: DLQMessageItem) {
    try {
      await resolveDLQMessage(item.id);
      await onRefreshTable(searchParams);
      toast.success("消息已标记为已解决");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function onPageChange(page_num: number) {
    try {
      await onRefreshTable({
        ...searchParams,
        page_num: page_num,
      });
      setSearchParams((prev) => ({
        ...prev,
        page_num: page_num,
      }));
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function onPageSizeChange(page_size: number) {
    try {
      await onRefreshTable({
        page_num: 1,
        page_size: page_size,
      });
      setSearchParams({
        page_num: 1,
        page_size: page_size,
      });
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  useEffect(() => {
    let ignore = false;

    if (user.role !== "admin") {
      router.replace("/dashboard");
    }

    async function fetchData() {
      try {
        const data = await getDLQMessages(searchParams);

        if (!ignore) {
          setData(data);
        }
      } catch (err) {
        const message =
          err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
        toast.error(message);
      }
    }

    fetchData();

    return () => {
      ignore = true;
    };
  }, [searchParams, router, user.role]);
  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MessageTable
        items={data.items}
        onRetry={handleRetry}
        onResolve={handleResolve}
      />
      <div className="mt-auto shrink-0">
        <Pagination
          pageNum={data.page_num}
          pageSize={data.page_size}
          total={data.total}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
        />
      </div>
    </div>
  );
}
