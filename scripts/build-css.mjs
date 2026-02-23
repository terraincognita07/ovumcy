import { spawnSync } from "node:child_process";

const buildResult = spawnSync(
  process.execPath,
  [
    "./node_modules/tailwindcss/lib/cli.js",
    "-i",
    "./web/src/css/input.css",
    "-o",
    "./web/static/css/tailwind.css",
    "--minify"
  ],
  {
    stdio: "inherit",
    env: {
      ...process.env,
      BROWSERSLIST_IGNORE_OLD_DATA: "1"
    }
  }
);

if (buildResult.error) {
  throw buildResult.error;
}

if (typeof buildResult.status === "number" && buildResult.status !== 0) {
  process.exit(buildResult.status);
}
