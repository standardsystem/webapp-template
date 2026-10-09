// Go のカバレッジがしきい値以上かを確認する。
//
//   node check-go-coverage.mjs <coverprofile> <しきい値(%)>
//
// CI とローカルで同じ判定を使うため、bc や awk に依存せず node で書いている。
import { execFileSync } from "node:child_process";
import { pathToFileURL } from "node:url";

// `go tool cover -func` の出力から total 行のカバレッジ（%）を取り出す。
export function parseTotal(coverFuncOutput) {
  const m = coverFuncOutput.match(/^total:\s+\(statements\)\s+([\d.]+)%\s*$/m);
  if (!m) {
    throw new Error("total line not found in `go tool cover -func` output");
  }
  return Number(m[1]);
}

function main(argv) {
  const [profile, thresholdArg] = argv;
  const threshold = Number(thresholdArg);
  if (!profile || !Number.isFinite(threshold)) {
    console.error("usage: node check-go-coverage.mjs <coverprofile> <threshold>");
    return 2;
  }

  const output = execFileSync("go", ["tool", "cover", `-func=${profile}`], {
    encoding: "utf8",
  });
  const total = parseTotal(output);

  if (total < threshold) {
    console.error(`Coverage ${total}% is below ${threshold}%`);
    return 1;
  }
  console.log(`Coverage ${total}% (threshold ${threshold}%)`);
  return 0;
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  process.exit(main(process.argv.slice(2)));
}
