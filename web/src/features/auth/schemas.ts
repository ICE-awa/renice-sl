import { z } from "zod";

export const loginSchema = z.object({
  identifier: z.string().min(1, "用户名或邮箱不可为空"),
  password: z
    .string()
    .min(8, "密码长度至少为 8 个字符")
    .max(72, "密码长度最多为 72 个字符"),
});

export const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, "用户名长度至少为 3 个字符")
      .max(20, "用户名长度最多为 20 个字符"),
    password: z
      .string()
      .min(8, "密码长度至少为 8 个字符")
      .max(72, "密码长度最多为 72 个字符"),
    confirm_password: z
      .string()
      .min(8, "确认密码长度至少为 8 个字符")
      .max(72, "确认密码长度最多为 72 个字符"),
    email: z.email("邮箱地址无效"),
    code: z.string().length(6, "请输入发送到您邮箱的 6 位验证码"),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: "密码不匹配",
    path: ["confirm_password"],
  });

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
