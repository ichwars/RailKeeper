import { fileURLToPath } from "node:url";
import typescript from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";

export default [
  { ignores: ["src/generated/**", "src/**/*.test.{ts,tsx}", "src/test/**"] },
  {
    files: ["src/**/*.{ts,tsx}"],
    languageOptions: {
      parser: typescript.parser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
        projectService: true,
        tsconfigRootDir: fileURLToPath(new URL("../../frontend/", import.meta.url))
      }
    },
    plugins: { "react-hooks": reactHooks, "@typescript-eslint": typescript.plugin },
    rules: {
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "error",
      "no-debugger": "error",
      "no-constant-binary-expression": "error",
      "no-unsafe-optional-chaining": "error",
      "@typescript-eslint/no-floating-promises": ["error", { ignoreVoid: false, ignoreIIFE: false }]
    }
  }
];
