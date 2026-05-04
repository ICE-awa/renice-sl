"use client";

import { useForm } from "react-hook-form";
import { type LoginInput, loginSchema } from "../schemas";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import { toast } from "sonner";
import { login } from "../api";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { scheduleRefresh } from "../../protected/components/session";

export function LoginForm() {
  const router = useRouter();

  const form = useForm<LoginInput>({
    resolver: standardSchemaResolver(loginSchema),
    defaultValues: {
      identifier: "",
      password: "",
    },
  });

  async function onSubmit(data: LoginInput) {
    try {
      const resp = await login(data);
      scheduleRefresh(resp.expires_in);
      toast.success("登录成功");
      router.push("/dashboard");
    } catch (err) {
      const message = err instanceof Error ? err.message : "登录失败";
      form.setError("root", { message });
      toast.error(message);
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>登录</CardTitle>
      </CardHeader>

      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.identifier}>
              <FieldLabel htmlFor="identifier">用户名或邮箱</FieldLabel>
              <Input
                id="identifier"
                placeholder="请输入用户名或邮箱"
                {...form.register("identifier")}
              />
              <FieldError errors={[form.formState.errors.identifier]} />
            </Field>
            <Field data-invalid={!!form.formState.errors.password}>
              <FieldLabel htmlFor="password">密码</FieldLabel>
              <Input
                id="password"
                type="password"
                {...form.register("password")}
              />
              <FieldDescription>请输入密码</FieldDescription>
              <FieldError errors={[form.formState.errors.password]} />
            </Field>
            <div className="space-y-3">
              <FieldError errors={[form.formState.errors.root]} />
              <Button
                type="submit"
                className="w-full"
                disabled={form.formState.isSubmitting}
              >
                {form.formState.isSubmitting ? "登录中..." : "登录"}
              </Button>
            </div>
          </FieldGroup>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          还没有账号？{" "}
          <Link
            href="/register"
            className="text-primary underline-offset-4 hover:underline"
          >
            去注册
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
