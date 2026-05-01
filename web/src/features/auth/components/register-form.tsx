"use client";
import { useForm } from "react-hook-form";
import { RegisterInput, registerSchema } from "../schemas";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
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
import { register } from "../api";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

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
      const message = err instanceof Error ? err.message : "注册失败";
      form.setError("root", { message });
      toast.error(message);
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>注册</CardHeader>

      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)}>
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
              <Input
                id="code"
                placeholder="请输入验证码"
                {...form.register("code")}
              />
              <FieldError errors={[form.formState.errors.code]} />
              {/* TODO 后续添加发送验证码的功能 */}
            </Field>
          </FieldGroup>
          <FieldError errors={[form.formState.errors.root]} />
          <Button
            type="submit"
            className="w-full"
            disabled={form.formState.isSubmitting}
          >
            注册
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
