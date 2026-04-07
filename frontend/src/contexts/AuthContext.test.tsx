import { render, screen, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { AuthProvider, useAuth } from "./AuthContext";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

function TestConsumer() {
  const { user, loading, logout } = useAuth();
  if (loading) return <div>loading</div>;
  if (!user) return <div>not authenticated</div>;
  return (
    <div>
      <span>user: {user.name}</span>
      <button onClick={logout}>logout</button>
    </div>
  );
}

describe("AuthContext", () => {
  it("正常系: 認証済みユーザーの情報を提供する", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: "u1",
          name: "Test",
          email: "t@e.com",
          avatarUrl: "",
          role: "member",
          createdAt: "",
        }),
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    expect(screen.getByText("loading")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText("user: Test")).toBeInTheDocument();
    });
  });

  it("異常系: 未認証時は null を返す", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "unauthorized" }),
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByText("not authenticated")).toBeInTheDocument();
    });
  });

  it("正常系: ログアウトで user が null になる", async () => {
    // /auth/me 成功
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: "u1",
          name: "Test",
          email: "t@e.com",
          avatarUrl: "",
          role: "member",
          createdAt: "",
        }),
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByText("user: Test")).toBeInTheDocument();
    });

    // /auth/logout 成功
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.reject(new Error("no body")),
    });

    const user = userEvent.setup();
    await act(async () => {
      await user.click(screen.getByText("logout"));
    });

    await waitFor(() => {
      expect(screen.getByText("not authenticated")).toBeInTheDocument();
    });
  });
});
