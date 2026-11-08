# Typed promise analysis compatibility

Checked on 2026-11-08. RailKeeper's frontend uses TypeScript 7.0.2 and React 19.3.1.
The isolated syntax/React-hook toolchain uses ESLint 10.14.0, Babel ESLint parser 8.0.7
and eslint-plugin-react-hooks 7.1.1. Type checking remains the actual strict project build;
syntax and hook checks remain mandatory for every production frontend source file.

The official [typescript-eslint support matrix](https://typescript-eslint.io/users/dependency-versions/)
currently supports TypeScript `>=4.8.4 <6.1.0`. ESLint 10 itself is supported, but the installed
TypeScript 7 compiler is outside the documented range. Open peer dependency ranges do not
establish support. Consequently `no-floating-promises` and `no-misused-promises` are not yet
introduced. No compiler downgrade, `--force`, unsupported-version warning suppression,
measurement exclusion or blanket `void` replacement is used to obtain a passing result.

Revisit this conditional item on the next TypeScript/typescript-eslint tool update, or by
2026-12-08. Before adoption, require all of the following:

1. Official support for the installed compiler and React/ESLint toolchain.
2. A separate pinned installation and reproducible clean `npm ci`.
3. The actual project type check, build and complete behavior/coverage suite.
4. Real error handling for findings, including rejected saves, cancellation and lost acknowledgements.
5. A fixture proving that each new rule rejects an unhandled promise and accepts the handled form.

This is the explicit compatibility deferral permitted by technical plan R7. It is not a claim
that type-aware promise rules ran successfully. Existing rules, strict compilation and complete
coverage remain active. Targeted behavior tests additionally verify configuration-save
acknowledgements and visible retryable exhibition failures without contacting physical devices.
