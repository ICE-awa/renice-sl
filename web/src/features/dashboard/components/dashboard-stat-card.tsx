"use client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type DashboardStatCardProps = {
  title: string;
  value: string | number;
  description?: string;
};

export default function DashboardStatCard({
  title,
  value,
  description,
}: DashboardStatCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>

      <CardContent className="flex">
        <div className="text-2xl font-semibold">{value}</div>
      </CardContent>
    </Card>
  );
}
