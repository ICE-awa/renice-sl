import { apiFetch } from "@/lib/api";
import { LoginInput, RegisterInput } from "./schemas";

type LoginResp = {
  expires_in: number;
};

export type RegisterConflictResp = {
  is_username_conflict: boolean;
  is_email_conflict: boolean;
};

export function login(input: LoginInput) {
  return apiFetch<LoginResp>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function register(input: RegisterInput) {
  const { confirm_password, ...payload } = input;

  return apiFetch<void>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
