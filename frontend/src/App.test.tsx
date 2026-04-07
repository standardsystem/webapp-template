import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import App from "./App";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
  window.history.pushState({}, "", "/");
});

describe("App", () => {
  it("未認証時はログインページを表示する", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "unauthorized" }),
    });

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText("ログインしてください")).toBeInTheDocument();
    });
  });

  it("未定義パスは / にリダイレクトされダッシュボードを表示する", async () => {
    window.history.pushState({}, "", "/unknown");
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () =>
          Promise.resolve({
            id: "u1",
            name: "Tester",
            email: "tester@example.com",
            avatarUrl: "",
            role: "member",
            createdAt: "",
          }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve([]),
      });

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText("Webapp Template")).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: "更新" })).toBeInTheDocument();
  });
});
