import parser from "@babel/eslint-parser";
import reactHooks from "eslint-plugin-react-hooks";

export default [
  { ignores: ["src/generated/**", "src/**/*.test.{ts,tsx}", "src/test/**"] },
  {
    files: ["src/**/*.{ts,tsx}"],
    languageOptions: {
      parser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
        requireConfigFile: false,
        babelOptions: {
          babelrc: false,
          configFile: false,
          parserOpts: { plugins: ["typescript", "jsx"] }
        }
      }
    },
    plugins: { "react-hooks": reactHooks },
    rules: {
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "error",
      "no-debugger": "error",
      "no-constant-binary-expression": "error",
      "no-unsafe-optional-chaining": "error"
    }
  }
];
