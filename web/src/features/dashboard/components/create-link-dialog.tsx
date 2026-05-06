"use client";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Controller, useForm } from "react-hook-form";
import { CreateLinkFormValues } from "../types";
import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import { createLinkSchema } from "../schemas";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { CalendarIcon } from "@phosphor-icons/react";
import { format } from "date-fns";
import { Calendar } from "@/components/ui/calendar";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";

type CreateLinkDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (values: {
    original_url: string;
    expires_at?: Date;
  }) => Promise<void> | void;
};

export default function CreateLinkDialog({
  open,
  onOpenChange,
  onSubmit,
}: CreateLinkDialogProps) {
  const form = useForm<CreateLinkFormValues>({
    resolver: standardSchemaResolver(createLinkSchema),
    defaultValues: {
      original_url: "",
      expires_at: undefined,
    },
  });

  async function handleSubmit(data: CreateLinkFormValues) {
    if (!data.original_url) {
      form.setError("original_url", { message: "请输入原链接" });
      return;
    }
    await onSubmit({
      original_url: data.original_url.trim(),
      expires_at: data.expires_at,
    });
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>新建短链接</DialogTitle>
        </DialogHeader>

        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.original_url}>
              <FieldLabel htmlFor="original_url">请填写原链接</FieldLabel>
              <Input
                {...form.register("original_url")}
                placeholder="请输入原链接"
                required
              />
              <FieldError errors={[form.formState.errors.original_url]} />
            </Field>

            <Field data-invalid={!!form.formState.errors.expires_at}>
              <FieldLabel htmlFor="expires_at">
                请选择过期时间（留空表示永不过期）
              </FieldLabel>
              <Controller
                control={form.control}
                name="expires_at"
                render={({ field }) => (
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        data-empty={field.value}
                        className="w-70 justify-start text-left font-normal data-[empty=true]:text-muted-foreground"
                      >
                        <CalendarIcon />
                        {field.value ? (
                          format(field.value, "PPP")
                        ) : (
                          <span>选择过期时间</span>
                        )}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0">
                      <Calendar
                        mode="single"
                        selected={field.value}
                        onSelect={field.onChange}
                      />
                    </PopoverContent>
                  </Popover>
                )}
              />
            </Field>
          </FieldGroup>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              取消
            </Button>

            <Button type="submit" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting ? "创建中..." : "创建"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
