import { Input } from "@/components/ui/input";
import DashboardStatCard from "@/features/dashboard/components/dashboard-stat-card";
import LinkTable from "@/features/dashboard/components/link-table";

export default function DashboardPage() {
  const links = [
    {
      id: 1,
      originalUrl: "https://verylongurl.com/verylong",
      shortUrl: "https://renice.cc/123456",
      viewCount: 123,
      createdAt: "2026-05-04 12:00:00",
      updatedAt: "2026-05-04 12:00:00",
      expiresAt: "2027-05-04 12:00:00",
    },
    {
      id: 2,
      originalUrl: "https://verylongurl.com/verylong",
      shortUrl: "https://renice.cc/123456",
      viewCount: 123,
      createdAt: "2026-05-04 12:00:00",
      updatedAt: "2026-05-04 12:00:00",
      expiresAt: "2027-05-04 12:00:00",
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
      </div>

      <div className="flex w-full [&>*]:flex-1 gap-12">
        <DashboardStatCard title="总链接数" value={2} />
        <DashboardStatCard title="总浏览量" value={123} />
      </div>

      <Input placeholder="搜索短链接 (TODO)" disabled />
      <LinkTable items={links} />
    </div>
  );
}
