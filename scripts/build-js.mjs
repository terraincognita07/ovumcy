import { copyFileSync } from "node:fs";

const buildTargets = [
  ["./node_modules/htmx.org/dist/htmx.min.js", "./web/static/js/htmx.min.js"],
  ["./node_modules/alpinejs/dist/cdn.min.js", "./web/static/js/alpine.min.js"],
  ["./web/src/js/app.js", "./web/static/js/app.js"],
  ["./web/src/js/settings-export.js", "./web/static/js/settings-export.js"]
];

for (const [source, destination] of buildTargets) {
  copyFileSync(source, destination);
}
