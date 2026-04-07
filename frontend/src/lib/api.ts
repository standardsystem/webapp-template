const API_BASE = "/api/v1";

type RequestOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
};

async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { method = "GET", body, headers = {} } = options;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error ?? "Unknown error");
  }

  // 204 No Content
  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

// --- 型定義 ---

export type AuthUser = {
  id: string;
  name: string;
  email: string;
  avatarUrl: string;
  role: "admin" | "member";
  createdAt: string;
};

export type User = {
  id: string;
  name: string;
  email: string;
  createdAt: string;
  updatedAt: string;
};

// --- Auth API ---

export const authApi = {
  me: () => request<AuthUser>("/auth/me"),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
  // login はバックエンドへのリダイレクトなので fetch ではなく window.location を使用
  loginURL: (provider: string) => `${API_BASE}/auth/${provider}/login`,
};

// --- User API ---

export const userApi = {
  list: () => request<User[]>("/users"),
  get: (id: string) => request<User>(`/users/${id}`),
  create: (input: { name: string; email: string }) =>
    request<User>("/users", { method: "POST", body: input }),
  delete: (id: string) => request<void>(`/users/${id}`, { method: "DELETE" }),
};

// --- Admin API ---

export const adminApi = {
  updateRole: (userId: string, role: "admin" | "member") =>
    request<void>(`/users/${userId}/role`, { method: "PUT", body: { role } }),
};

export const healthApi = {
  check: async (): Promise<{ status: string; timestamp: string }> => {
    const res = await fetch("/health");
    if (!res.ok) {
      throw new Error(`Health check failed: ${res.statusText}`);
    }
    return res.json();
  },
};
