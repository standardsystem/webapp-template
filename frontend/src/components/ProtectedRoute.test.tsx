import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ProtectedRoute } from "./ProtectedRoute";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

function renderWithProviders(
  ui: React.ReactElement,
  { initialEntries = ["/"] } = {},
) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AuthProvider>{ui}</AuthProvider>
    </MemoryRouter>,
  );
}

const authenticatedUser = {
  id: "u1",
  name: "Test",
  email: "t@e.com",
  avatarUrl: "",
  role: "member" as const,
  createdAt: "",
};

const adminUser = { ...authenticatedUser, role: "admin" as const };

describe("ProtectedRoute", () => {
  it("ローディング中は読み込み表示", () => {
    // fetch を永遠に pending にする
    mockFetch.mockReturnValueOnce(new Promise(() => {}));

    renderWithProviders(
      <ProtectedRoute>
        <div>protected content</div>
      </ProtectedRoute>,
    );

    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("認証済みユーザーは子コンポーネントを表示", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(authenticatedUser),
    });

    renderWithProviders(
      <ProtectedRoute>
        <div>protected content</div>
      </ProtectedRoute>,
    );

    await waitFor(() => {
      expect(screen.getByText("protected content")).toBeInTheDocument();
    });
  });

  it("未認証ユーザーは /login にリダイレクト", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "unauthorized" }),
    });

    renderWithProviders(
      <ProtectedRoute>
        <div>protected content</div>
      </ProtectedRoute>,
    );

    await waitFor(() => {
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  it("admin 必須ルートで member はアクセス拒否", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(authenticatedUser),
    });

    renderWithProviders(
      <ProtectedRoute requiredRole="admin">
        <div>admin content</div>
      </ProtectedRoute>,
    );

    await waitFor(() => {
      expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
    });
    expect(screen.queryByText("admin content")).not.toBeInTheDocument();
  });

  it("admin ユーザーは admin ルートにアクセスできる", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(adminUser),
    });

    renderWithProviders(
      <ProtectedRoute requiredRole="admin">
        <div>admin content</div>
      </ProtectedRoute>,
    );

    await waitFor(() => {
      expect(screen.getByText("admin content")).toBeInTheDocument();
    });
  });
});
