import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { DashboardPage } from "./DashboardPage";

const logoutMock = vi.fn();
const refetchMock = vi.fn();

vi.mock("@/contexts/AuthContext", () => ({
  useAuth: () => ({
    user: { id: "u1", name: "Tester", role: "member" },
    logout: logoutMock,
  }),
}));

vi.mock("@/hooks/useUsers", () => ({
  useUsers: () => ({
    users: [{ id: "u1", name: "Tester", email: "t@example.com", createdAt: "", updatedAt: "" }],
    loading: false,
    error: null,
    refetch: refetchMock,
  }),
}));

describe("DashboardPage", () => {
  beforeEach(() => {
    logoutMock.mockReset();
    refetchMock.mockReset();
  });

  it("ユーザー情報と一覧を表示する", () => {
    render(<DashboardPage />);

    expect(screen.getByText("Tester（member）")).toBeInTheDocument();
    expect(screen.getByText("t@example.com")).toBeInTheDocument();
  });

  it("更新ボタンとログアウトボタンが動作する", async () => {
    const user = userEvent.setup();
    render(<DashboardPage />);

    await user.click(screen.getByRole("button", { name: "更新" }));
    await user.click(screen.getByRole("button", { name: "ログアウト" }));

    expect(refetchMock).toHaveBeenCalledTimes(1);
    expect(logoutMock).toHaveBeenCalledTimes(1);
  });
});
