import { copyFileSync } from "node:fs";

copyFileSync("./node_modules/htmx.org/dist/htmx.min.js", "./web/static/js/htmx.min.js");
copyFileSync("./node_modules/alpinejs/dist/cdn.min.js", "./web/static/js/alpine.min.js");
copyFileSync("./web/src/js/app.js", "./web/static/js/app.js");
copyFileSync("./web/src/js/settings-export.js", "./web/static/js/settings-export.js");
