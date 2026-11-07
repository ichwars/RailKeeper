import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { ESLint } from "eslint";

const frontend = fileURLToPath(new URL("../../frontend/", import.meta.url));
const eslint = new ESLint({
  cwd: frontend,
  overrideConfigFile: fileURLToPath(new URL("./eslint.config.mjs", import.meta.url))
});
const filePath = "src/quality-lint-fixture.tsx";

test("TypeScript/JSX still detects conditional hooks and stale dependencies", async () => {
  const [result] = await eslint.lintText(`
    import { useEffect, useState } from "react";
    export function Example({ enabled, value }: { enabled: boolean; value: string }) {
      if (enabled) useState(0);
      useEffect(() => { document.title = value; }, []);
      return <span>{value}</span>;
    }
  `, { filePath });
  const rules = result.messages.map(message => message.ruleId);
  assert.ok(rules.includes("react-hooks/rules-of-hooks"), JSON.stringify(result.messages));
  assert.ok(rules.includes("react-hooks/exhaustive-deps"), JSON.stringify(result.messages));
  assert.ok(result.messages.every(message => message.ruleId), "fixture must parse successfully");
});

test("valid typed effects and real Effect Event triggers remain accepted", async () => {
  const [result] = await eslint.lintText(`
    import { useEffect, useEffectEvent } from "react";
    export function Example({ owner, value }: { owner: string; value: string }) {
      const onOwnerEntry = useEffectEvent(() => { document.title = value; });
      useEffect(() => { onOwnerEntry(); }, [owner]);
      return <span>{value}</span>;
    }
  `, { filePath });
  assert.deepEqual(result.messages, []);
});

test("production files are checked while generated contracts and test fixtures are excluded", async () => {
  assert.equal(await eslint.isPathIgnored(fileURLToPath(new URL("../../frontend/src/generated/api.d.ts", import.meta.url))), true);
  assert.equal(await eslint.isPathIgnored(fileURLToPath(new URL("../../frontend/src/shared/i18n.test.ts", import.meta.url))), true);
  assert.equal(await eslint.isPathIgnored(fileURLToPath(new URL("../../frontend/src/shared/i18n.ts", import.meta.url))), false);
});
