import { apiFetch } from "@/lib/api/client";
import type { AuthUser } from "@/lib/stores/auth.store";

type ApiUser = {
  id: string;
  name: string;
  email: string;
  phone?: string;
  platform_role: "user" | "super_admin";
};

type ApiAuthResponse = {
  user: ApiUser;
  access_token: string;
  refresh_token: string;
};

type ApiMeResponse = {
  user: ApiUser;
  tenants: unknown[];
};

export type AuthSession = {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
};

export type RegisterInput = {
  name: string;
  email: string;
  phone?: string;
  password: string;
};

export type LoginInput = {
  email: string;
  password: string;
};

export async function register(input: RegisterInput): Promise<AuthSession> {
  const response = await apiFetch<ApiAuthResponse>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(input),
    tenantScoped: false
  });

  return normalizeSession(response);
}

export async function login(input: LoginInput): Promise<AuthSession> {
  const response = await apiFetch<ApiAuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
    tenantScoped: false
  });

  return normalizeSession(response);
}

export async function me() {
  const response = await apiFetch<ApiMeResponse>("/api/v1/auth/me", {
    tenantScoped: false
  });

  return {
    user: normalizeUser(response.user)
  };
}

export async function logout(refreshToken: string) {
  await apiFetch<void>("/api/v1/auth/logout", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
    tenantScoped: false
  });
}

function normalizeSession(response: ApiAuthResponse): AuthSession {
  return {
    user: normalizeUser(response.user),
    accessToken: response.access_token,
    refreshToken: response.refresh_token
  };
}

function normalizeUser(user: ApiUser): AuthUser {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    phone: user.phone,
    platformRole: user.platform_role
  };
}
