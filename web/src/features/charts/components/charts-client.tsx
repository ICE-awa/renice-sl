"use client";

import { useEffect, useState } from "react";
import { ChartItem } from "./chart-item";
import { Switch } from "@/components/ui/switch";
import { getClickStat, getLinkStat, getUserStat } from "../api";
import { StatResponse } from "../type";
import { ApiError } from "@/lib/api";
import { toast } from "sonner";
import { useSession } from "@/features/protected/components/session-provider";
import { useRouter } from "next/navigation";

export function ChartsClient() {
  const [granularity, setGranularity] = useState<"day" | "hour">("day");
  const [clickData, setClickData] = useState([] as StatResponse[]);
  const [userData, setUserData] = useState([] as StatResponse[]);
  const [linkData, setLinkData] = useState([] as StatResponse[]);
  const { user } = useSession();
  const router = useRouter();

  function handleGranularityChange(value: "hour" | "day") {
    setGranularity(value);
  }

  useEffect(() => {
    let ignore = false;

    if (user.role !== "admin") {
      router.replace("/dashboard");
      return;
    }

    async function initialChartData() {
      if (ignore) return;
      try {
        await getClickStat({
          range: granularity === "hour" ? 24 : 7,
          bucket: granularity,
        }).then((data) => setClickData(data));

        await getUserStat({
          range: granularity === "hour" ? 24 : 7,
          bucket: granularity,
        }).then((data) => setUserData(data));

        await getLinkStat({
          range: granularity === "hour" ? 24 : 7,
          bucket: granularity,
        }).then((data) => setLinkData(data));
      } catch (err) {
        const message =
          err instanceof ApiError ? err.message : "服务器打了个盹，请稍后再试";
        toast.error(message);
      }
    }

    void initialChartData();

    return () => {
      ignore = true;
    };
  }, [granularity, router, user.role]);

  return (
    <div className="flex flex-1 flex-col min-h-0 gap-4">
      <div className="flex shrink-0 justify-center items-center py-4 px-4">
        <h1 className="text-2xl font-semibold ml-auto">状态图表</h1>

        <div className="flex items-center gap-2 text-sm text-muted-foreground ml-auto">
          <span>日</span>
          <Switch
            checked={granularity === "hour"}
            onCheckedChange={(checked) =>
              handleGranularityChange(checked ? "hour" : "day")
            }
          />
          <span>小时</span>
        </div>
      </div>

      <div className="grid min-h-0 flex-1 grid-cols-2 grid-rows-2 gap-4 py-4 px-4">
        <ChartItem data={clickData} granularity={granularity} label="点击量" />
        <ChartItem
          data={userData}
          granularity={granularity}
          label="用户新增量"
        />
        <ChartItem
          data={linkData}
          granularity={granularity}
          label="链接新增量"
        />
      </div>
    </div>
  );
}
