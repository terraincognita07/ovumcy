import { copyFileSync } from "node:fs";

copyFileSync("./node_modules/htmx.org/dist/htmx.min.js", "./web/static/js/htmx.min.js");
copyFileSync("./node_modules/alpinejs/dist/cdn.min.js", "./web/static/js/alpine.min.js");
