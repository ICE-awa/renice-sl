import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";

type StatPoint = {
  time: string;
  count: number;
};

type ChartItemProps = {
  data: StatPoint[];
  granularity: "hour" | "day";
  label: string;
  color?: string;
};

function formatTime(value: string, granularity: "hour" | "day") {
  const date = new Date(value);

  if (granularity === "hour") {
    return new Intl.DateTimeFormat("zh-CN", {
      timeZone: "Asia/Shanghai",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).format(date);
  }

  return new Intl.DateTimeFormat("zh-CN", {
    timeZone: "Asia/Shanghai",
    month: "2-digit",
    day: "2-digit",
  })
    .format(date)
    .replace(/\//g, "-");
}

export function ChartItem({
  data,
  granularity,
  label,
  color = "var(--chart-1)",
}: ChartItemProps) {
  const chartConfig = {
    count: {
      label,
      color,
    },
  } satisfies ChartConfig;

  return (
    <div>
      <ChartContainer config={chartConfig} className="h-70 w-full">
        <LineChart accessibilityLayer data={data}>
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="time"
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            tickFormatter={(value) => formatTime(String(value), granularity)}
          />
          <YAxis tickLine={false} axisLine={false} allowDecimals={false} />
          <ChartTooltip
            content={<ChartTooltipContent />}
            labelFormatter={(value) => formatTime(String(value), granularity)}
          />
          <Line
            type="monotone"
            dataKey="count"
            stroke="var(--color-count)"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ChartContainer>

      <div className="text-center mt-3 text-sm text-muted-foreground">
        {label}
        {granularity === "hour" ? " (按小时) " : " (按天) "}趋势
      </div>
    </div>
  );
}
