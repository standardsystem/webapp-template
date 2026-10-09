import assert from "node:assert/strict";
import { test } from "node:test";

import { parseTotal } from "./check-go-coverage.mjs";

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
