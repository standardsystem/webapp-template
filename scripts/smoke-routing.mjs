// フロントエンドから API への経路のスモークテスト。
//
//   node scripts/smoke-routing.mjs
//
// 次の 2 つの経路で、/health と /api がバックエンドに届くことを確認する。
//   1. 開発時: Docker Compose の frontend（Vite の開発サーバー）→ backend
//   2. 本番時: frontend の本番イメージ（nginx）→ backend
//
// OAuth ログインは含まない（プロバイダのクライアントが必要なため）。
// 開発用の環境とは別の compose プロジェクトで起動するが、ポート 5173・8080・5432 を使うので、
// `mise run dev` を止めてから実行すること。
import { execFileSync } from "node:child_process";

const PROJECT = "webapp-smoke";
const DEV_ORIGIN = "http://localhost:5173";
const PROD_PORT = 8081;
const PROD_ORIGIN = `http://localhost:${PROD_PORT}`;
const PROD_CONTAINER = `${PROJECT}-frontend-prod`;
const PROD_IMAGE = "webapp-frontend";

function docker(args, { quiet = false } = {}) {
  return execFileSync("docker", args, {
    encoding: "utf8",
    stdio: ["ignore", "pipe", quiet ? "ignore" : "inherit"],
  }).trim();
}

const compose = (args, opts) => docker(["compose", "-p", PROJECT, ...args], opts);

async function waitFor(description, check, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  let lastError;
  while (Date.now() < deadline) {
    try {
      await check();
      console.log(`ok: ${description}`);
      return;
    } catch (err) {
      lastError = err;
      await new Promise((resolve) => setTimeout(resolve, 2000));
    }
  }
  throw new Error(`timed out waiting for: ${description} (${lastError?.message})`);
}

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

// /health がバックエンドの JSON を返すこと。
async function checkHealth(origin) {
  const res = await fetch(`${origin}/health`);
  assert(res.status === 200, `GET /health: status ${res.status}`);
  const body = await res.json();
  assert(body.status === "ok", `GET /health: body ${JSON.stringify(body)}`);
}

// 未ログインの /api/v1/auth/me が、index.html ではなくバックエンドの 401（JSON）を返すこと。
async function checkApiReachesBackend(origin) {
  const res = await fetch(`${origin}/api/v1/auth/me`);
  const text = await res.text();
  assert(res.status === 401, `GET /api/v1/auth/me: status ${res.status}, body ${text.slice(0, 200)}`);
  assert(
    (res.headers.get("content-type") ?? "").includes("application/json"),
    `GET /api/v1/auth/me: content-type ${res.headers.get("content-type")}`,
  );
  assert(JSON.parse(text).error, `GET /api/v1/auth/me: body ${text.slice(0, 200)}`);
}

// API 以外のパスは SPA の index.html を返すこと。
async function checkSpaFallback(origin) {
  for (const path of ["/", "/some/client/route"]) {
    const res = await fetch(`${origin}${path}`);
    const text = await res.text();
    assert(res.status === 200, `GET ${path}: status ${res.status}`);
    assert(text.includes('<div id="root">'), `GET ${path}: not index.html`);
  }
}

function cleanup() {
  try {
    docker(["rm", "-f", PROD_CONTAINER], { quiet: true });
  } catch {
    // コンテナがなければ何もしない
  }
  // -p で分けたプロジェクトのコンテナとボリュームだけを消す（開発用のデータには触れない）
  compose(["down", "-v", "--remove-orphans"]);
}

async function main() {
  try {
    console.log("== 1. Docker Compose（Vite の開発サーバー → backend）");
    compose(["up", "-d", "--build", "db", "migrate", "backend", "frontend"]);
    // backend は go run でコンパイルしてから起動するため、初回は時間がかかる
    await waitFor("dev: GET /health via Vite proxy", () => checkHealth(DEV_ORIGIN), 300_000);
    await checkApiReachesBackend(DEV_ORIGIN);
    console.log("ok: dev: GET /api/v1/auth/me reaches the backend (401 JSON)");

    console.log("== 2. 本番イメージ（nginx → backend）");
    const backendId = compose(["ps", "-q", "backend"]);
    const network = docker([
      "inspect",
      "-f",
      "{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{end}}",
      backendId,
    ]);
    docker([
      "run", "-d", "--name", PROD_CONTAINER,
      "--network", network,
      "-e", "BACKEND_URL=http://backend:8080",
      "-p", `${PROD_PORT}:8080`,
      PROD_IMAGE,
    ]);
    await waitFor("prod: GET /health via nginx", () => checkHealth(PROD_ORIGIN), 60_000);
    await checkApiReachesBackend(PROD_ORIGIN);
    console.log("ok: prod: GET /api/v1/auth/me reaches the backend (401 JSON)");
    await checkSpaFallback(PROD_ORIGIN);
    console.log("ok: prod: other paths return index.html");
  } catch (err) {
    console.error(`FAILED: ${err.message}`);
    try {
      console.error(compose(["logs", "--tail", "40", "backend", "frontend", "migrate"]));
      console.error(docker(["logs", "--tail", "40", PROD_CONTAINER], { quiet: true }));
    } catch {
      // ログが取れなくても元の失敗を報告する
    }
    process.exitCode = 1;
  } finally {
    cleanup();
  }
}

await main();
