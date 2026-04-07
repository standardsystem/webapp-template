import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { LoginPage } from "./LoginPage";

describe("LoginPage", () => {
  it("OAuth ボタンを表示する", () => {
    render(<LoginPage />);

    expect(screen.getByRole("button", { name: "Google でログイン" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "GitHub でログイン" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Microsoft でログイン" })).toBeInTheDocument();
  });

});
