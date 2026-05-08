"use client";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { LinkItem, UpdateLinkFormValues, UpdateLinkInput } from "../types";
import { Controller, useForm } from "react-hook-form";
import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import { updateLinkSchema } from "../schemas";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { CalendarIcon } from "@phosphor-icons/react";
import { format } from "date-fns";
import { Calendar } from "@/components/ui/calendar";
import { Switch } from "@/components/ui/switch";
import { useEffect } from "react";

type EditLinkDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (values: UpdateLinkFormValues) => Promise<void> | void;
  item: LinkItem | null;
};

export default function EditLinkDialog({
  open,
  onOpenChange,
  onSubmit,
  item,
}: EditLinkDialogProps) {
  const form = useForm<UpdateLinkFormValues>({
    resolver: standardSchemaResolver(updateLinkSchema),
    defaultValues: {
      original_url: "",
      expires_at: undefined,
      enabled: true,
      no_expires: false,
    },
  });

  async function handleSubmit(data: UpdateLinkFormValues) {
    if (item === null) {
      return;
    }

    try {
      await onSubmit({
        id: item.id,
        original_url: data.original_url?.trim(),
        expires_at: data.expires_at,
        enabled: data.enabled,
        no_expires: data.expires_at === undefined ? true : false,
      });
      onOpenChange(false);
    } catch {}
  }

  useEffect(() => {
    if (item === null) {
      return;
    }
    form.reset({
      original_url: item.original_url,
      expires_at: item.expires_at ? new Date(item.expires_at) : undefined,
      enabled: item.status === "active" ? true : false,
      no_expires: item.expires_at ? false : true,
    });
  }, [item, form]);

  if (item === null) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑短链接</DialogTitle>
        </DialogHeader>

        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.original_url}>
              <FieldLabel htmlFor="original_url">原链接</FieldLabel>
              <Input
                {...form.register("original_url")}
                placeholder={item.original_url}
              />
              <FieldError>
                {form.formState.errors.original_url?.message}
              </FieldError>
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
                      >
                        <CalendarIcon />
                        {field.value
                          ? format(field.value, "PPP")
                          : "请选择过期时间"}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0">
                      <Calendar
                        mode="single"
                        selected={field.value}
                        onSelect={(date) => {
                          field.onChange(date);

                          if (date) {
                            form.setValue("no_expires", false, {
                              shouldDirty: true,
                              shouldValidate: true,
                            });
                          }
                        }}
                      />
                    </PopoverContent>
                  </Popover>
                )}
              />
            </Field>

            <Field data-invalid={!!form.formState.errors.enabled}>
              <FieldLabel htmlFor="status">状态</FieldLabel>
              <Controller
                control={form.control}
                name="enabled"
                render={({ field }) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
            </Field>

            <Field data-invalid={!!form.formState.errors.no_expires}>
              <FieldLabel htmlFor="no_expires">永不过期</FieldLabel>
              <Controller
                control={form.control}
                name="no_expires"
                render={({ field }) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={(checked) => {
                      field.onChange(checked);

                      if (checked) {
                        form.setValue("expires_at", undefined, {
                          shouldDirty: true,
                          shouldValidate: true,
                        });
                      }
                    }}
                  />
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
              {form.formState.isSubmitting ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
