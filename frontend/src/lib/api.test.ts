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

  it("user create/delete と admin updateRole を呼び出せる", async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () =>
          Promise.resolve({
            id: "u1",
            name: "Alice",
            email: "alice@example.com",
            createdAt: "",
            updatedAt: "",
          }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.reject(new Error("no body")),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.reject(new Error("no body")),
      });

    const created = await userApi.create({ name: "Alice", email: "alice@example.com" });
    expect(created.id).toBe("u1");

    await expect(userApi.delete("u1")).resolves.toBeUndefined();
    await expect(adminApi.updateRole("u1", "admin")).resolves.toBeUndefined();
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
