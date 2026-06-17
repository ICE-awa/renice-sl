"use client";
import { useForm } from "react-hook-form";
import { RegisterInput, registerSchema } from "../schemas";
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
import { refresh, register } from "../api";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import Link from "next/link";
import { RegisterConflictResp } from "../types";
import { useEffect } from "react";

export function RegisterForm() {
  const router = useRouter();

  const form = useForm<RegisterInput>({
    resolver: standardSchemaResolver(registerSchema),
    defaultValues: {
      username: "",
      password: "",
      confirm_password: "",
      email: "",
      code: "",
    },
  });

  async function onSubmit(data: RegisterInput) {
    try {
      await register(data);
      toast.success("注册成功");
      router.push("/login");
    } catch (err) {
      if (err instanceof ApiError) {
        const conflict = err.data as RegisterConflictResp | undefined;

        if (conflict?.is_username_conflict) {
          form.setError("username", { message: "用户名已被占用" });
        }

        if (conflict?.is_email_conflict) {
          form.setError("email", { message: "邮箱已被占用" });
        }

        if (conflict?.is_username_conflict || conflict?.is_email_conflict) {
          toast.error("邮箱名或用户已被占用");
          return;
        }
      }
      toast.error("请检查表单项是否正确填写");
    }
  }

  async function onSendCode() {
    // TODO 后续添加发送验证码的功能
  }

  useEffect(() => {
    let alive = true;

    refresh()
      .then(() => {
        if (!alive) return;
        router.replace("/dashboard");
      })
      .catch(() => {});

    return () => {
      alive = false;
    };
  }, [router]);

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>注册</CardTitle>
      </CardHeader>

      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.username}>
              <FieldLabel htmlFor="username">用户名</FieldLabel>
              <Input
                id="username"
                placeholder="请输入用户名"
                {...form.register("username")}
              />
              <FieldDescription>
                请输入用户名，3-20 字符，只能使用英文和数字
              </FieldDescription>
              <FieldError errors={[form.formState.errors.username]} />
            </Field>
            <Field data-invalid={!!form.formState.errors.password}>
              <FieldLabel htmlFor="password">密码</FieldLabel>
              <Input
                id="password"
                type="password"
                {...form.register("password")}
              />
              <FieldDescription>
                请输入密码，8-32 字符，可以使用英文，数字以及特殊字符
              </FieldDescription>
              <FieldError errors={[form.formState.errors.password]} />
            </Field>
            <Field data-invalid={!!form.formState.errors.confirm_password}>
              <FieldLabel htmlFor="confirm_password">确认密码</FieldLabel>
              <Input
                id="confirm_password"
                type="password"
                {...form.register("confirm_password")}
              />
              <FieldError errors={[form.formState.errors.confirm_password]} />
            </Field>
            <Field data-invalid={!!form.formState.errors.email}>
              <FieldLabel htmlFor="email">邮箱</FieldLabel>
              <Input
                id="email"
                type="email"
                placeholder="请输入邮箱"
                {...form.register("email")}
              />
              <FieldError errors={[form.formState.errors.email]} />
            </Field>
            <Field data-invalid={!!form.formState.errors.code}>
              <FieldLabel htmlFor="code">验证码</FieldLabel>
              <div className="flex gap-2">
                <Input
                  id="code"
                  placeholder="请输入验证码"
                  {...form.register("code")}
                />

                <Button type="button" variant="outline" onClick={onSendCode}>
                  获取邮箱验证码
                </Button>
                {/* TODO 后续添加发送验证码的功能 */}
              </div>
              <FieldError errors={[form.formState.errors.code]} />
            </Field>
          </FieldGroup>
          <div className="space-y-3">
            <FieldError errors={[form.formState.errors.root]} />
            <Button
              type="submit"
              className="w-full"
              disabled={form.formState.isSubmitting}
            >
              {form.formState.isSubmitting ? "注册中..." : "注册"}
            </Button>
          </div>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          已经有账号？{" "}
          <Link
            href="/login"
            className="text-primary underline-offset-4 hover:underline"
          >
            去登录
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
