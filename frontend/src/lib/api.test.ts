import { describe, expect, it, vi, beforeEach } from "vitest";
import { adminApi, authApi, healthApi, userApi } from "./api";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("api helpers", () => {
  it("auth loginURL は正しいパスを返す", () => {
    expect(authApi.loginURL("google")).toBe("/api/v1/auth/google/login");
  });

  it("user get は GET /users/:id を呼び出す", async () => {
    const alice = {
      id: "u1",
      name: "Alice",
      email: "alice@example.com",
      createdAt: "",
      updatedAt: "",
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(alice),
    });

    await expect(userApi.get("u1")).resolves.toEqual(alice);
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/v1/users/u1",
      expect.objectContaining({ method: "GET", credentials: "include" }),
    );
  });

  it("admin updateRole は PUT /users/:id/role を呼び出し、204 で undefined を返す", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.reject(new Error("no body")),
    });

    await expect(adminApi.updateRole("u1", "admin")).resolves.toBeUndefined();
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/v1/users/u1/role",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ role: "admin" }),
      }),
    );
  });

  it("エラー応答は本文の error を message にして投げる", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      statusText: "Not Found",
      json: () => Promise.resolve({ error: "not found" }),
    });

    await expect(userApi.get("missing")).rejects.toThrow("not found");
  });

  it("health check 成功時にレスポンスを返す", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ status: "ok", timestamp: "now" }),
    });

    await expect(healthApi.check()).resolves.toEqual({ status: "ok", timestamp: "now" });
  });

  it("health check 失敗時はエラーにする", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
    });

    await expect(healthApi.check()).rejects.toThrow("Health check failed: Internal Server Error");
  });
});
