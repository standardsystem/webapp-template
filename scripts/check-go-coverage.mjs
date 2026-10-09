// Go のカバレッジがしきい値以上かを確認する。
//
//   node check-go-coverage.mjs <coverprofile> <しきい値(%)> [--exclude=<ディレクトリ>,<ディレクトリ>...]
//
// --exclude に指定したディレクトリ（例: internal/mock,cmd）以下のファイルは、計測対象から外す。
// CI とローカルで同じ判定を使うため、bc や awk に依存せず node で書いている。
import { execFileSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

// `go tool cover -func` の出力から total 行のカバレッジ（%）を取り出す。
export function parseTotal(coverFuncOutput) {
  const m = coverFuncOutput.match(/^total:\s+\(statements\)\s+([\d.]+)%\s*$/m);
  if (!m) {
    throw new Error("total line not found in `go tool cover -func` output");
  }
  return Number(m[1]);
}

// coverprofile から、excludes のディレクトリ以下にあるファイルの行を取り除く。先頭の mode 行は残す。
export function filterProfile(profileText, excludes) {
  return profileText
    .split("\n")
    .filter((line) => {
      if (line.startsWith("mode:")) return true;
      // 各行は "<パッケージパス>/<ファイル>.go:<開始>,<終了> <文の数> <実行回数>"
      const file = line.slice(0, line.indexOf(".go:") + 3);
      return !excludes.some((dir) => file.includes(`/${dir}/`));
    })
    .join("\n");
}

function main(argv) {
  const positional = argv.filter((a) => !a.startsWith("--"));
  const excludeArg = argv.find((a) => a.startsWith("--exclude="));
  const excludes = excludeArg
    ? excludeArg.slice("--exclude=".length).split(",").filter(Boolean)
    : [];

  const [profile, thresholdArg] = positional;
  const threshold = Number(thresholdArg);
  if (!profile || !Number.isFinite(threshold)) {
    console.error(
      "usage: node check-go-coverage.mjs <coverprofile> <threshold> [--exclude=a,b]",
    );
    return 2;
  }

  let target = profile;
  if (excludes.length > 0) {
    target = `${profile}.filtered`;
    writeFileSync(target, filterProfile(readFileSync(profile, "utf8"), excludes));
  }

  const output = execFileSync("go", ["tool", "cover", `-func=${target}`], {
    encoding: "utf8",
  });
  const total = parseTotal(output);
  const scope = excludes.length > 0 ? `, excluding ${excludes.join(", ")}` : "";

  if (total < threshold) {
    console.error(`Coverage ${total}% is below ${threshold}%${scope}`);
    return 1;
  }
  console.log(`Coverage ${total}% (threshold ${threshold}%${scope})`);
  return 0;
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  process.exit(main(process.argv.slice(2)));
}
