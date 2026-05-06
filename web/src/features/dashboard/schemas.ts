import z from "zod";

export const createLinkSchema = z
  .object({
    original_url: z.url("请输入有效的链接"),
    expires_at: z.date().optional(),
  })
  .refine((data) => {
    if (!data.expires_at) return true;

    const expires_at = new Date(data.expires_at);
    const now = new Date();
    return expires_at > now;
  });

export const updateLinkSchema = z
  .object({
    original_url: z.url("请输入有效的链接").optional(),
    expires_at: z.date().optional(),
    enabled: z.boolean(),
  })
  .refine((data) => {
    if (!data.expires_at) return true;

    const expires_at = new Date(data.expires_at);
    const now = new Date();
    return expires_at > now;
  });

export type CreateLinkInput = z.infer<typeof createLinkSchema>;
export type UpdateLinkInput = z.infer<typeof updateLinkSchema>;
