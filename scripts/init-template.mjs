// テンプレートから作成したリポジトリの名前を、案件用に一括で書き換える。
//
//   mise run init -- <org>/<repo> [--title "画面に表示する名前"]
//
// 書き換えるもの（手順の全体は docs/getting-started.md）:
//   - Go のモジュールパス  github.com/your-org/webapp-template → github.com/<org>/<repo>
//   - リソース名の接頭辞    webapp-template（Cloud Run のサービス名・ジョブ名、JWT の issuer、package.json の name）→ <repo>
//   - CLI のバイナリ名      webapp-cli → <repo>-cli（cli/cmd/webapp-cli ディレクトリも改名する）
//   - 画面に表示する名前    Webapp Template → --title の値（省略時は <repo>）
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, renameSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const TEMPLATE_ORG = "your-org";
const TEMPLATE_REPO = "webapp-template";
const TEMPLATE_CLI = "webapp-cli";
const TEMPLATE_TITLE = "Webapp Template";
const TEMPLATE_MODULE = `github.com/${TEMPLATE_ORG}/${TEMPLATE_REPO}`;

// 書き換えないファイル。ロックファイルと、テンプレートの名前そのものを説明しているファイル。
const SKIP_FILES = new Set([
  "frontend/pnpm-lock.yaml",
  "backend/go.sum",
  "go.work.sum",
  "docs/getting-started.md",
  "scripts/init-template.mjs",
  "scripts/init-template.test.mjs",
]);

// GitHub のオーナー名とリポジトリ名に使える文字だけを受け付ける。
const NAME_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]*$/;

export function parseArgs(argv) {
  const positional = [];
  let title;
  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === "--title") {
      title = argv[++i];
    } else if (argv[i].startsWith("--title=")) {
      title = argv[i].slice("--title=".length);
    } else {
      positional.push(argv[i]);
    }
  }
  if (positional.length !== 1) {
    throw new Error('usage: mise run init -- <org>/<repo> [--title "Display Name"]');
  }
  const parts = positional[0].split("/");
  if (parts.length !== 2 || !parts.every((p) => NAME_PATTERN.test(p))) {
    throw new Error(`"${positional[0]}" is not in the form <org>/<repo>`);
  }
  const [org, repo] = parts;
  if (title !== undefined && title.trim() === "") {
    throw new Error("--title must not be empty");
  }
  return { org, repo, title: title ?? repo };
}

// 1 回の走査で置換する。置換後の文字列が別のパターンに再び一致することはない
// （リポジトリ名に "webapp-template" を含む場合でも二重に置換しない）。
export function rewrite(text, { org, repo, title }) {
  const replacements = new Map([
    [TEMPLATE_MODULE, `github.com/${org}/${repo}`],
    [TEMPLATE_REPO, repo],
    [TEMPLATE_CLI, `${repo}-cli`],
    [TEMPLATE_TITLE, title],
  ]);
  // 長いパターンを先に並べ、モジュールパスが接頭辞の置換より優先されるようにする
  const pattern = new RegExp(
    [...replacements.keys()]
      .sort((a, b) => b.length - a.length)
      .map((k) => k.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"))
      .join("|"),
    "g",
  );
  return text.replace(pattern, (match) => replacements.get(match));
}

// NUL を含むファイルはバイナリとして扱い、書き換えない。
export function isBinary(buffer) {
  return buffer.includes(0);
}

function main(argv) {
  let names;
  try {
    names = parseArgs(argv);
  } catch (err) {
    console.error(err.message);
    return 2;
  }

  if (!readFileSync("backend/go.mod", "utf8").includes(TEMPLATE_MODULE)) {
    console.error(`backend/go.mod does not contain ${TEMPLATE_MODULE}; this repository is already initialized.`);
    return 1;
  }

  const files = execFileSync("git", ["ls-files", "-z"], { encoding: "utf8" })
    .split("\0")
    .filter((f) => f && !SKIP_FILES.has(f) && existsSync(f));

  let changed = 0;
  for (const file of files) {
    const buffer = readFileSync(file);
    if (isBinary(buffer)) continue;
    const before = buffer.toString("utf8");
    const after = rewrite(before, names);
    if (after !== before) {
      writeFileSync(file, after);
      changed++;
    }
  }

  const cliDir = `cli/cmd/${TEMPLATE_CLI}`;
  const newCliDir = `cli/cmd/${names.repo}-cli`;
  if (existsSync(cliDir)) {
    renameSync(cliDir, newCliDir);
  }

  console.log(`Rewrote ${changed} files for github.com/${names.org}/${names.repo}.`);
  console.log(`Renamed ${cliDir} to ${newCliDir}.`);
  console.log("Next: mise run fmt && mise run check (see docs/getting-started.md).");
  return 0;
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  process.exit(main(process.argv.slice(2)));
}
