# Typed promise analysis compatibility

The production build remains on TypeScript 7.0.2 with strict checking. The isolated
`tools/frontend-lint` installation uses ESLint 10.14.0, typescript-eslint 8.72.1,
eslint-plugin-react-hooks 7.1.1 and the supported TypeScript 6 compiler API (resolved
version 6.0.3 in the lockfile).
Its `typescript` package is pinned to `npm:@typescript/typescript6@6.0.2`.
`frontend/package.json` installs both toolchains reproducibly with `npm ci`.

Microsoft explicitly recommends a side-by-side TypeScript 6 API for tools using the
compiler API in the [TypeScript 7 announcement](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
This API version is within the official
[typescript-eslint support matrix](https://typescript-eslint.io/users/dependency-versions/).
The application compiler is neither downgraded nor passed to an unsupported parser.
No `--force`, warning suppression or narrowed production scope is used.

`@typescript-eslint/no-floating-promises` is mandatory for every production TypeScript
and TSX file alongside the existing syntax and React hook rules. `ignoreVoid: false`
and `ignoreIIFE: false` mean that explicitly discarding a promise cannot hide a rejection.
The parser's project service uses the actual frontend `tsconfig.json` and its React types.

Actions retain their existing asynchronous return contracts and local error/draft handling.
The shared `observeAction` boundary observes explicitly backgrounded work and reports an
unexpected rejection through the browser's existing error channel. It does not convert
asynchronous callbacks into synchronous callbacks or duplicate handled errors. A failed
logout leaves the authenticated workspace available for an explicit retry. Sidebar save
queue cleanup handles both outcomes without creating an unobserved rejected promise.

The lint fixtures prove that conditional hooks, stale dependencies, a bare promise and
`void promise` are rejected. Correctly handled promises and valid Effect Event usage pass.
Behavior tests additionally cover rejected saves, cancellation, retained drafts, logout
retry and session ownership. Strict compilation and the complete coverage suite remain
required. These automated checks do not establish physical command-station compatibility.
