"use client";
import { Input } from "@/components/ui/input";
import { GetLinksInput } from "../types";
import { useState } from "react";
import { Button } from "@/components/ui/button";

type LinkSearchBarProps = {
  onSearch: (params: GetLinksInput) => Promise<void> | void;
  onCreate: () => void;
  onReset: () => void;
};

export default function LinkSearchBar({
  onSearch,
  onReset,
  onCreate,
}: LinkSearchBarProps) {
  const [originalUrl, setOriginalUrl] = useState("");
  const [code, setCode] = useState("");

  async function handleSearch() {
    await onSearch({
      original_url: originalUrl.trim(),
      code: code.trim(),
      page_num: 1,
      page_size: 10,
    });
  }

  function handleReset() {
    setOriginalUrl("");
    setCode("");
    onReset();
  }

  return (
    <div className="grid grid-cols-[20%_20%_10%_10%_10%_10%_10%_10%] items-center gap-3">
      <Input
        value={originalUrl}
        onChange={(event) => setOriginalUrl(event.target.value)}
        placeholder="搜索原链接"
        className="col-span-1"
      />

      <Input
        value={code}
        onChange={(event) => setCode(event.target.value)}
        placeholder="搜索短链接后缀"
        className="col-span-1"
      />
      <div className="col-span-5 flex justify-end gap-2">
        <Button type="button" onClick={handleSearch}>
          搜索
        </Button>
        <Button type="button" variant="outline" onClick={handleReset}>
          重置
        </Button>
        <Button type="button" onClick={onCreate}>
          新建短链接
        </Button>
      </div>
    </div>
  );
}
