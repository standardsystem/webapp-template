import assert from "node:assert/strict";
import { test } from "node:test";

import { filterProfile, parseTotal } from "./check-go-coverage.mjs";

test("parseTotal は total 行のカバレッジを返す", () => {
  const output = [
    "example.com/app/internal/handler/health.go:15:\tHealth\t\t100.0%",
    "example.com/app/internal/usecase/user.go:20:\tGetUser\t\t75.0%",
    "total:\t\t\t\t\t\t\t(statements)\t27.2%",
    "",
  ].join("\n");

  assert.equal(parseTotal(output), 27.2);
});

test("parseTotal は関数名に total を含む行を拾わない", () => {
  const output = [
    "example.com/app/internal/usecase/stats.go:10:\ttotal\t\t0.0%",
    "total:\t\t\t\t\t\t\t(statements)\t100.0%",
  ].join("\n");

  assert.equal(parseTotal(output), 100);
});

test("parseTotal は total 行がなければエラーにする", () => {
  assert.throws(() => parseTotal("no coverage here\n"), /total line not found/);
});

test("filterProfile は指定したディレクトリ以下の行を取り除き、mode 行は残す", () => {
  const profile = [
    "mode: atomic",
    "example.com/app/cmd/server/main.go:25.13,30.2 3 0",
    "example.com/app/internal/mock/user_repository.go:20.1,22.2 1 0",
    "example.com/app/internal/handler/health.go:15.1,23.2 5 1",
    "example.com/app/internal/usecase/cmdline.go:10.1,12.2 1 1",
    "",
  ].join("\n");

  const filtered = filterProfile(profile, ["cmd", "internal/mock"]);

  assert.equal(
    filtered,
    [
      "mode: atomic",
      "example.com/app/internal/handler/health.go:15.1,23.2 5 1",
      "example.com/app/internal/usecase/cmdline.go:10.1,12.2 1 1",
      "",
    ].join("\n"),
  );
});

test("filterProfile は除外指定がなければ何も取り除かない", () => {
  const profile = "mode: set\nexample.com/app/cmd/server/main.go:25.13,30.2 3 0\n";

  assert.equal(filterProfile(profile, []), profile);
});
