import js from "@eslint/js";
import globals from "globals";

export default [
  {
    ignores: [
      "node_modules/",
      "web/src/js/app/*.js",
      "web/static/js/alpine.min.js",
      "web/static/js/htmx.min.js"
    ]
  },
  {
    files: ["web/src/js/**/*.js", "web/static/js/chart-lite.js"],
    ...js.configs.recommended,
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "script",
      globals: {
        ...globals.browser,
        htmx: "readonly"
      }
    }
  },
  {
    files: ["scripts/*.mjs"],
    ...js.configs.recommended,
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: globals.node
    }
  }
];
