# Reproducible quality checks

## Compiler and scope

`.go-version` pins Go 1.27.3 for CI runtime tests, CodeQL, native Windows qualification and
release builds. Docker uses the same version. `check_go_toolchain.py` rejects drift between
these settings. `backend/go.mod` separately declares the supported minimum, Go 1.26.10;
a dedicated CI job runs the complete suite with `GOTOOLCHAIN=local` at that minimum.

The pinned golangci-lint 2.11.4 analyzer runs against the minimum compiler. Its current
export-data reader cannot decode Go 1.27's version-5 export data. Using the declared
minimum retains every existing lint rule while both compilers run the actual application
suite. This is an explicit analysis-tool limitation, not a claim that lint ran under Go 1.27.

## Frontend

Vitest includes every `src/**/*.{ts,tsx}` file. Only tests, test fixtures, declarations and
the generated API declaration file are excluded. An unimported production module therefore
continues to affect the result. CI requires at least 80% statements and lines and 75%
branches and functions. These percentages never replace the critical behavioral cases.

The active exhibition entry point delegates to `ExhibitionWorkspacePage`. Its previous
non-exported `LegacyExhibitionView` had no caller and was removed, including its private
helpers. The active route and existing exhibition capabilities remain. No measurement
exclusion was added to remove that unreachable implementation from coverage.

Behavior tests cover rejected and explicitly retried saves, retained drafts, unknown physical
dates, invalid dates, source selection, viewer boundaries, second-factor/recovery flows,
conflict confirmation and report/QR output. Date selection uses the chosen DE/EN language
while retaining ISO and day.month.year input. Opening a calendar never substitutes today.

```sh
npm --prefix frontend run test:coverage
npm --prefix frontend run lint
npm --prefix frontend run build
```

## Backend

The Linux backend profile instruments all `internal` packages together, including code
reached through API or application integration tests. A package's individual test-run
percentage is not its final coverage. The merged statement profile is the measurement.
`check_go_coverage.py` merges repeated blocks from different test binaries once, using any
observed execution. Its result agrees with `go tool cover`; it rejects missing required
modules and coverage below `tools/go-coverage-baseline.json`.

The measured starting result for this scope was 74.73%. The total floor is 74%, with
separate floors for every current internal module, including authentication, credentials,
decoder/CV, backup, transfer, valuation and their storage adapters. Those module floors
are the rounded-down measured results. The subsequently added `buildinfo` module has a
measured 100% statement floor. They are regression guards, not a declaration that
uncovered branches or physical-device behavior have been qualified.

```sh
go -C backend test -coverpkg=./internal/... -coverprofile=coverage.out ./...
go -C backend tool cover -func=coverage.out
python3 tools/check_go_coverage.py --profile backend/coverage.out --report backend/coverage-summary.json
python3 -m unittest discover -s tools -p test_go_coverage.py
```

The existing required **Backend tests** check succeeds only after coverage, the minimum-compiler
suite, every race slice and bounded fuzz have all succeeded. Its final aggregation runs even
when a dependency fails or is skipped and rejects those results.

CI also runs separate race-instrumented rule/storage, authentication, transfer/ECoS, backup
and decoder/CV groups. The decoder group targets configuration, drafts, imports, sources,
files, binding conflicts and atomicity. Exhaustive bundled decoder-catalog combinations
remain in the complete non-race suite; they are not mislabeled as concurrency tests.
Each race slice has a 20-minute Go timeout and a 30-minute job limit to accommodate actual
instrumented SQLite work. No test assertion or role boundary is disabled.

Bounded fuzzing runs for password-hash parsing, quoted protocol comparison names, immutable
CV partial-state validation and backup path/content integrity. Inputs are synthetic and
bounded; no real command station or external account is contacted.

```sh
go -C backend test ./internal/backup -run '^$' -fuzz '^FuzzBackupFileValidation$' -fuzztime=10s -parallel=2
```

Windows CI retains only qualification reports and screenshots as diagnostic artifacts,
including on failure. Runtime databases, original backup data, authentication files and
launch credentials are excluded. Native package qualification still tests the actual
resource-patched executable before an archive is published.

Full preservation, browser, platform and human acceptance remain governed by the
[testing baseline](testing-baseline.md) and [pilot/rollback procedure](modernization-rollout.md).
