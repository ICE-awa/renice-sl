import { apiFetch } from "@/lib/api";
import { LoginInput, RegisterInput } from "./schemas";
import { CurrentUser, LoginResp } from "./types";

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

export function refresh() {
  return apiFetch<LoginResp>("/api/v1/auth/refresh", {
    method: "POST",
  });
}

export function logout() {
  return apiFetch<void>("/api/v1/auth/logout", {
    method: "POST",
  });
}

export function me() {
  return apiFetch<CurrentUser>("/api/v1/auth/me", {
    method: "GET",
  });
}
