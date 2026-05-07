"use client";
import { useEffect, useState } from "react";
import { createLink, deleteLink, getLinks, getStats, updateLink } from "../api";
import {
  CreateLinkFormValues,
  GetLinksInput,
  GetStatsResponse,
  LinkItem,
  UpdateLinkFormValues,
} from "../types";
import DashboardStatCard from "./dashboard-stat-card";
import LinkTable from "./link-table";
import { ApiError } from "@/lib/api";
import { toast } from "sonner";
import LinkSearchBar from "./link-search-bar";
import CreateLinkDialog from "./create-link-dialog";
import EditLinkDialog from "./edit-link-dialog";

const originalSearchParams: GetLinksInput = {
  page_num: 1,
  page_size: 10,
};

export default function DashboardClient() {
  const [links, setLinks] = useState<LinkItem[]>([]);
  const [stats, setStats] = useState<GetStatsResponse>({
    link_count: 0,
    view_count: 0,
  });
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editingLink, setEditingLink] = useState<LinkItem | null>(null);

  const [searchParams, setSearchParams] =
    useState<GetLinksInput>(originalSearchParams);

  async function onRefreshTable(params: GetLinksInput) {
    const data = await getLinks(params);
    const stats = await getStats();
    setStats(stats);
    setLinks(data);
  }

  async function handleSearch(params: GetLinksInput) {
    try {
      const data = await getLinks(params);
      setSearchParams(params);
      setLinks(data);
      toast.info(`搜索完成，共找到 ${data.length} 条结果`);
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function handleCreate({
    original_url,
    expires_at,
  }: CreateLinkFormValues) {
    try {
      await createLink({ original_url, expires_at: expires_at?.toISOString() });
      //   const params: GetLinksInput = {
      //     page_num: 1,
      //     page_size: 10,
      //   };
      //   await handleSearch(params);
      setCreateOpen(false);
      toast.success("新建短链接成功");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
      throw err;
    }

    try {
      await onRefreshTable(searchParams);
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function handleReset() {
    try {
      setSearchParams(originalSearchParams);
      await onRefreshTable(originalSearchParams);
      toast.info("成功重置搜索条件");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function handleStatusChange(id: number, status: "active" | "inactive") {
    try {
      await updateLink({
        id: id,
        status: status === "active" ? "inactive" : "active",
      });
      await onRefreshTable(searchParams);
      toast.success("状态已更新");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      const code = err instanceof ApiError ? err.code : 0;
      console.error(code);
      toast.error(message);
    }
  }

  function handleOnEdit(link: LinkItem) {
    setEditingLink(link);
    setEditOpen(true);
  }

  async function handleEdit(values: UpdateLinkFormValues) {
    try {
      await updateLink({
        id: values.id,
        original_url: values.original_url,
        expires_at: values.expires_at
          ? values.expires_at.toISOString()
          : undefined,
        status: values.enabled ? "active" : "inactive",
      });
      toast.success("成功更新链接信息");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
      throw err;
    }

    try {
      await onRefreshTable(searchParams);
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  async function handleDelete(link: LinkItem) {
    try {
      await deleteLink(link.id);
      await onRefreshTable(searchParams);
      toast.success("成功删除链接");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
      toast.error(message);
    }
  }

  useEffect(() => {
    let ignore = false;

    async function loadInitialLinks() {
      try {
        const data = await getLinks(originalSearchParams);
        const stats = await getStats();

        if (!ignore) {
          setStats(stats);
          setLinks(data);
        }
      } catch (err) {
        const message =
          err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
        toast.error(message);
      }
    }

    void loadInitialLinks();

    return () => {
      ignore = true;
    };
  }, []);

  //   useEffect(() => {
  //     const params: GetLinksInput = {
  //         page_num: 1,
  //         page_size: 10,
  //     }
  //     handleSearch(params)
  //   }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
      </div>

      <div className="space-y-4">
        <div className="flex w-full gap-6 [&>*]:flex-1">
          <DashboardStatCard title="总链接数" value={stats.link_count} />
          <DashboardStatCard title="总浏览量" value={stats.view_count} />
        </div>
        <LinkSearchBar
          onSearch={handleSearch}
          onCreate={() => {
            setCreateOpen(true);
          }}
          onReset={handleReset}
        />
        <CreateLinkDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          onSubmit={handleCreate}
        />
        <LinkTable
          items={links}
          onEdit={handleOnEdit}
          onStatusChange={handleStatusChange}
          onDelete={handleDelete}
        />
        <EditLinkDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          onSubmit={handleEdit}
          item={editingLink}
        />
      </div>
    </div>
  );
}
