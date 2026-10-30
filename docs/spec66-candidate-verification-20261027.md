# SPEC66 candidate verification — 2026-10-27

## Decision and immutable input

**Resource-selected race verification completed successfully. Build, six-family non-race tests and six-family vet passed. Repository-wide vet remains FAILED. This is not full SPEC66 acceptance or release approval.**

- Repository: `williamxhero/CLIProxyAPI-ex`.
- Candidate branch: `claude/spec66-resource-catalog`.
- Candidate HEAD: `3e1488e4f1c6256aea14c7c4a0542bd475f38761`.
- Candidate tree: `f10442be6e2152eb41ab6eebd51779fa11b7fada`.
- Parent: `a4acc9f752bd46571f737a10c04bf413656ab06b`.
- Coordination SPEC: <https://github.com/williamxhero/CPA-Manager-Plus-ex/issues/66> (read as OPEN in this run).
- Checks below actually ran in the original `CLIProxyAPI-source` checkout against this candidate on Windows/amd64 with `go1.27.0`, `CGO_ENABLED=1`, `CC=gcc`.
- No implementation, fixture, crypto cost, assertion, account, credential, route or production configuration was changed. No merge, deployment, production-8317 request or manual execution of the server EXE occurred. Existing tests, including their own test-process helpers, ran through `go test` only.

The original checkout already had an uncommitted change to `docs/resource-catalog-v1-verification.md`. It was read, never written, staged or restored. Its pre-receipt/final-integrity reference SHA256 is `d50875ef21262b5a45cfcd3b916852f9aa58637d93f2e46a6f83ffb37f11e55f`. This new SPEC receipt deliberately leaves that historical document untouched. The candidate branch/ref was not changed. This receipt is isolated on `claude/spec66-verification-receipt-20261027`, based exactly on the candidate.

An initial EnterWorktree call failed because the session started at the non-repository parent with no WorktreeCreate hook. Checks continued without source edits. After the session directory moved to the actual repository, an attempted standalone receipt write was rejected by the isolation guard before any file was changed. A second EnterWorktree succeeded, initially based on origin/main; a new documentation branch was then explicitly created at the tested candidate SHA before writing this file. **Checks are receipts for the original candidate checkout, not a claim that tests were rerun in the documentation worktree.** `git diff --exit-code HEAD -- '*.go' go.mod go.sum` returned 0 after checks in the original checkout. No source patch or merge was used to overcome the test timeout.

## Environment and evidence provenance

Go commands ran from the actual module with these process-local settings:

```powershell
$S = 'C:\Users\will\.claude\jobs\4b1d02f8\tmp\spec66-verify'
$env:GOCACHE = "$S\cache"
$env:GOTMPDIR = "$S\build"
$env:GOPROXY = 'off'
$env:GOSUMDB = 'off'
$env:GOTOOLCHAIN = 'local'
```

Cache/build subdirectories were created before successful runs. GOMODCACHE was inherited at `C:\Users\will\go\pkg\mod`. No cache relocation or global Go/settings/ACL change was made. No task-only telemetry-OFF mechanism was configured in this run; historical telemetry-OFF receipts must not be attributed to these commands.

Commands, exits and package-result lines below are transcribed from actual tool outputs in this session, not imported as fresh results from earlier issues. They are not byte-for-byte copies of separately retained raw stdout/stderr files. No unrecorded wall time, per-subtest count or per-command timestamp is inferred. Complete output files for the first and successful third race runs were retained by the session at these temporary paths:

```text
%LOCALAPPDATA%\Temp\claude\D--WILL-AGENT-CPA\4b1d02f8-7d85-474d-9a11-2d354eb37d91\tasks\b86kpjeuf.output
%LOCALAPPDATA%\Temp\claude\D--WILL-AGENT-CPA\4b1d02f8-7d85-474d-9a11-2d354eb37d91\tasks\bu4awwooq.output
```

The second race output was truncated by the tool; its visible timeout is preserved below, not presented as a complete saved raw log. This document durably retains successful result lines and command/coverage distinctions; the coordination issue receives a separate summary and readback.

Two initial setup errors preceded successful verification: running Go from the non-module parent returned `go.mod file not found`; missing GOTMPDIR then returned `The system cannot find the file specified`. Neither ran tests or constitutes a candidate assertion failure. Changing to the actual module and creating scratch subdirectories resolved those errors. No permission-denial workaround was needed for Go verification in this run.

## 1. Required server build — PASS

Actual Go invocation, with `$S` as above:

```powershell
go build -o (Join-Path $S 'cli-proxy-api-check.exe') ./cmd/server
```

Exit **0**, no build diagnostic printed. Output exists, **94,130,176 bytes**, SHA256:

```text
11612be8e39061b27231a34cc9643a6705524691ad23893bba5a4d3a3eebb14f
```

The server EXE was not executed. It remains disposable scratch output, not deployment/runtime evidence. Historical build permission diagnostics are not erased by this separate clean build result.

## 2. Six related package families — PASS

Exact invocation:

```text
go test ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/... -count=1 -timeout=300s
```

Actual exit **0**. The common import prefix below is `github.com/router-for-me/CLIProxyAPI/v8/`:

```text
ok  internal/resourcecatalog             43.922s
ok  internal/config                      43.201s
ok  internal/watcher                     44.422s
ok  internal/watcher/diff                42.065s
ok  internal/watcher/synthesizer         43.346s
ok  internal/api                        110.328s
ok  internal/api/handlers/management     50.188s
ok  internal/api/middleware              41.744s
ok  sdk/auth                            43.404s
ok  sdk/cliproxy                         49.555s
ok  sdk/cliproxy/auth                    47.668s
ok  sdk/cliproxy/executionregistry       40.672s
ok  sdk/cliproxy/executor                15.384s
?   sdk/cliproxy/pipeline                [no test files]
?   sdk/cliproxy/resource                [no test files]
ok  sdk/cliproxy/session                 20.122s
ok  sdk/cliproxy/usage                   14.885s
```

Current `go list` of all six recursive family patterns resolved **17 packages: 15 passing package results and 2 packages with no test files**, not 17 tested packages. `internal/config` and `sdk/auth` have no descendant packages here, so their exact selectors cover their entire current family. Historical receipts with different package totals are not silently reused. This was not `go test ./...`.

## 3. Vet — scoped PASS; repository-wide FAIL

Exact six-family command:

```text
go vet ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

Actual exit **0**, no diagnostics.

An additional separate command was actually run:

```text
go vet ./...
```

Actual exit **1**, with these **16** diagnostics:

```text
internal/pluginhost/loader_windows.go:302:31: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:314:30: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:368:32: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:375:20: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:386:34: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:401:28: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:402:31: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows.go:421:17: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows_test.go:53:28: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows_test.go:54:31: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows_test.go:142:30: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows_test.go:153:28: possible misuse of unsafe.Pointer
internal/pluginhost/loader_windows_test.go:154:31: possible misuse of unsafe.Pointer
internal/runtime/executor/kimi_executor.go:1389:9: ForAPIKey passes lock by value: KimiExecutor contains ClaudeExecutor contains sync.Mutex
internal/runtime/executor/kimi_executor.go:1391:21: assignment copies lock value to e.ClaudeExecutor: ClaudeExecutor contains sync.Mutex
internal/runtime/executor/oauth_scope_executor.go:12:9: ForAPIKey passes lock by value: ClaudeExecutor contains sync.Mutex
```

Type import prefixes in the last three lines are abbreviated; actual diagnostics identify types in `github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor`. The candidate's parent-to-HEAD diff for these four files is empty. **No baseline vet execution was performed**: unchanged-file inspection is not a reproduced baseline comparison. No unrelated production fix was made. Repository-wide vet remains an open gate.

## 4. API family non-race verification — separate PASS

```text
go test ./internal/api/... -count=1 -timeout=600s
```

Actual exit **0**:

```text
ok  github.com/router-for-me/CLIProxyAPI/v8/internal/api                      81.285s
ok  github.com/router-for-me/CLIProxyAPI/v8/internal/api/handlers/management   9.750s
ok  github.com/router-for-me/CLIProxyAPI/v8/internal/api/middleware            2.990s
```

This is full current API-family **non-race** coverage, not API full-family race coverage. It did not time out. It is distinct from the Resource-only race timeouts below.

## 5. Resource race — two timeouts retained; complete selected coverage now PASS

The focused non-race prerequisite passed first:

```text
go test ./internal/api ./internal/resourcecatalog -run '^TestResource' -count=1 -timeout=120s
ok  github.com/router-for-me/CLIProxyAPI/v8/internal/api              66.911s
ok  github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog   3.792s
```

Its exit was **0**. It is not race evidence.

### Actual race attempts, in order

1. **TIMEOUT / exit 1**:

   ```text
   go test -race ./internal/api ./internal/resourcecatalog -run '^TestResource' -count=1 -timeout=240s
   panic: test timed out after 4m0s
   running tests: TestResourceInventoryInheritedManualClears/duplicate=true
   FAIL github.com/router-for-me/CLIProxyAPI/v8/internal/api 242.918s
   ok   github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog 4.158s
   ```

   The active stack entered `bcrypt.GenerateFromPassword` from `config.ParseConfigBytes` through `newInventoryHarness`. No data-race warning was present in the retained output. The tool moved this command to the background after a 300-second outer wait and subsequently reported actual completed exit 1. This is a Go test timeout, not a pass or a still-running invocation.

2. **TIMEOUT / exit 1**:

   ```text
   go test -race ./internal/api ./internal/resourcecatalog -run '^TestResource' -count=1 -timeout=540s
   panic: test timed out after 9m0s
   running tests: TestResourceInventoryManualJSONRoundTrip/gemini
   ```

   The tool output was truncated; no complete raw log or additional package-result detail is claimed here. This remains a failed aggregate invocation.

3. **PASS / exit 0**, with the original Resource selector retained in one API test process:

   ```text
   go test -race -parallel=4 ./internal/api ./internal/resourcecatalog -run '^TestResource' -count=1 -timeout=900s
   ok   github.com/router-for-me/CLIProxyAPI/v8/internal/api             775.911s
   ok   github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog   4.169s
   EXIT_CODE=0
   ```

   This completed after being moved to the background by the tool's 600-second outer wait; its finished output was read back. No data-race warning appeared. All 55 API top-level Resource names were selected in **one test process**, unlike earlier per-test-isolated receipts. Catalog contributes five selected top-level names.

   The successful command **increased the test-time budget to 900s** and specified `-parallel=4`; it is not equivalent to passing earlier 240/540/600-second gates. Source inspection found no `t.Parallel()` in Resource test files. A prior progress-message suggestion that reducing package-internal parallelism fixed contention is withdrawn. No causal parallelism fix, crypto-cost fix, lifecycle fix or source repair is established. Bcrypt accumulation is an observed plausible cost path, not proven as the sole cause; shared-process cleanup concerns are not disproven.

### Remaining candidate-family Resource coverage — PASS with explicit empty selections

```text
go test -race ./internal/config/... ./internal/watcher/... ./internal/api/handlers/management ./internal/api/middleware ./sdk/auth/... ./sdk/cliproxy/... -run '^TestResource' -count=1 -timeout=180s
```

Actual exit **0**. Watcher successfully executed its Resource selection:

```text
ok github.com/router-for-me/CLIProxyAPI/v8/internal/watcher 44.427s
```

The other **12 test-bearing packages** in this supplemental command reported `[no tests to run]`, not full-package race passes:

```text
internal/config; internal/watcher/diff; internal/watcher/synthesizer;
internal/api/handlers/management; internal/api/middleware; sdk/auth;
sdk/cliproxy; sdk/cliproxy/auth; sdk/cliproxy/executionregistry;
sdk/cliproxy/executor; sdk/cliproxy/session; sdk/cliproxy/usage
```

`sdk/cliproxy/pipeline` and `sdk/cliproxy/resource` reported `[no test files]`.

### Manifest and coverage differences

Actual enumeration also exited 0:

```text
go test ./internal/api ./internal/resourcecatalog ./internal/watcher -list '^TestResource'
```

It enumerated **55 API + 5 catalog + 1 watcher = 61 top-level test names**. Enumeration is not execution; selected-package execution is evidenced by the successful race commands above. Quiet output does not establish a new per-subtest PASS/SKIP count.

- The six families cover the candidate's **11 changed Go-package directories**, including no-test `sdk/cliproxy/resource`, plus six related directories.
- The supplemental race run closes the watcher omission in the first two-package race command. Empty selectors in other packages do not supply full race coverage.
- The two successful race commands together select all `^TestResource` cases in the six families without per-API-test isolation. This is two package-group commands, **not one all-six-family invocation**.
- These are **Resource-only** race receipts, not full SDK/management/middleware/config race suites or repository-wide race acceptance.
- `sdk/pluginapi` has an unrelated Resource-named test but is outside both the six requested families and candidate-changed packages. It is not included or silently counted.
- Full `go test -race ./internal/api/...` **was not run anew in this task**. Historical full-family race timeouts remain failures, not converted to passes by Resource selection or the fresh full-API non-race pass.
- No fresh repository-wide `go test ./...`, baseline comparison, Linux runtime/durability, power-loss, rollback/rebind, UI, deployment or production acceptance is claimed.

## Conclusion and handoff

The requested Resource-selected race check now has a real successful same-process API receipt, plus catalog and watcher coverage. The immediate aggregate time-budget blocker was cleared operationally by allowing 900 seconds; that does not satisfy an unchanged shorter deadline or establish a code defect/fix. The required build and all current six-family non-race package tests passed; scoped vet passed. Full-repository vet remains red with the 16 diagnostics above.

No merge or deployment is authorized by these results. Keep SPEC66 open: repository-wide vet, any required full-package API race gate, baseline attribution and the previously documented broader acceptance boundaries remain unresolved. This new owner-project SPEC receipt is isolated from the candidate's pre-existing dirty historical document. Commit/push and issue publication are not yet performed; each is limited to this receipt and its separate documentation-only branch.
