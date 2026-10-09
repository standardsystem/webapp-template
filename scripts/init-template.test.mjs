import assert from "node:assert/strict";
import { test } from "node:test";

import { isBinary, parseArgs, rewrite } from "./init-template.mjs";

test("parseArgs は <org>/<repo> を分け、title の既定値を repo にする", () => {
  assert.deepEqual(parseArgs(["acme/orders"]), { org: "acme", repo: "orders", title: "orders" });
});

test("parseArgs は --title を 2 通りの書き方で受け付ける", () => {
  assert.equal(parseArgs(["acme/orders", "--title", "受注管理"]).title, "受注管理");
  assert.equal(parseArgs(["--title=Order Desk", "acme/orders"]).title, "Order Desk");
});

test("parseArgs は不正な引数をエラーにする", () => {
  for (const argv of [
    [],
    ["orders"],
    ["acme/orders/extra"],
    ["acme/"],
    ["/orders"],
    ["acme/or ders"],
    ["acme/orders; rm -rf ."],
    ["acme/orders", "other/repo"],
    ["acme/orders", "--title", " "],
  ]) {
    assert.throws(() => parseArgs(argv), Error, JSON.stringify(argv));
  }
});

const names = { org: "acme", repo: "orders", title: "受注管理" };

test("rewrite は Go のモジュールパスを書き換える", () => {
  assert.equal(
    rewrite('import "github.com/your-org/webapp-template/internal/domain"\n', names),
    'import "github.com/acme/orders/internal/domain"\n',
  );
  assert.equal(
    rewrite("module github.com/your-org/webapp-template/cli\n", names),
    "module github.com/acme/orders/cli\n",
  );
});

test("rewrite はリソース名・CLI 名・表示名を書き換える", () => {
  const before = [
    "BACKEND_SERVICE: webapp-template-api",
    "MIGRATE_JOB: webapp-template-migrate",
    'issuer:    "webapp-template",',
    '"name": "webapp-template-frontend",',
    "go build -o webapp-cli ./cmd/webapp-cli",
    "<title>Webapp Template</title>",
  ].join("\n");

  assert.equal(
    rewrite(before, names),
    [
      "BACKEND_SERVICE: orders-api",
      "MIGRATE_JOB: orders-migrate",
      'issuer:    "orders",',
      '"name": "orders-frontend",',
      "go build -o orders-cli ./cmd/orders-cli",
      "<title>受注管理</title>",
    ].join("\n"),
  );
});

test("rewrite はテンプレートと関係のない webapp を書き換えない", () => {
  const text = "POSTGRES_DB: webapp\ndocker build -t webapp-backend .\n/webapp/${{ env.BACKEND_SERVICE }}\n";

  assert.equal(rewrite(text, names), text);
});

test("rewrite は置換後の文字列を二重に置換しない", () => {
  const nested = { org: "your-org", repo: "my-webapp-template", title: "My Webapp Template" };

  assert.equal(
    rewrite("github.com/your-org/webapp-template webapp-template Webapp Template", nested),
    "github.com/your-org/my-webapp-template my-webapp-template My Webapp Template",
  );
});

test("isBinary は NUL を含むデータだけをバイナリと判定する", () => {
  assert.equal(isBinary(Buffer.from("plain text\n")), false);
  assert.equal(isBinary(Buffer.from([0x89, 0x50, 0x00, 0x47])), true);
});
