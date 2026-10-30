# Resource catalog v1: implementation and verification record

Date: October 26, 2026 (Asia/Singapore). **Uncommitted prerequisite slice; not full
T01/SPEC66/SPEC79 completion.** The original worker stopped at an actual tool/cache
permission diagnostic, without retrying or bypassing it. No nested agents or model substitute
were used. This record is local owner-project review material, not a deployment
or external publication. The original-worker receipts below are historical;
see **Pass 5: owned saves, endpoint clears and effective credential seals** for
the current frozen input and unstaged corrections. Earlier sections retain their historical snapshots and failures. The previously denied server build is still blocked.

## Branch and scope

- Branch: `claude/spec66-resource-catalog`.
- Base and unchanged HEAD: `a4acc9f752bd46571f737a10c04bf413656ab06b`.
- The original worker began with a clean prepared tree and left its changes
  unstaged/uncommitted. The coordinator subsequently staged the 44-file candidate
  for immutable review. After the original follow-up, the coordinator staged the
  complete 48-file candidate described in Pass 4, then the 49-file candidate
  described in Pass 5. Only the Pass 5 corrections are now unstaged; the older
  44/48-file statuses below are historical.
- No branch/ref/index mutation, fetch, rebase/reset, commit, push, PR/comment,
  service startup, deployment, task card, or schedule operation was performed.
- Only the owning source workspace was changed. The manager repository was not
  changed. Runtime configs, `.env`, personal auth material and live databases/
  providers were not inspected or called.

## Delivered behavior and compatibility

The source implements opt-in standalone file-backed resource identity, complete
raw-source inventory, additive authenticated v8 reads, source marker carriage,
version/CAS fences, tombstones, conservative ambiguity, lifetime writer locking,
and recoverable individually durable publication. It preserves native v8 groups
through legacy typed saves, including filtered source rows. GET has no migration,
refresh, upstream call or gate side effect. Source/config/storage conflicts return
generic unavailable responses rather than stale complete inventory.

Default opt-out is unchanged. Enabling is an explicit source-migration decision;
state defaults beside the chosen config and outside the auth tree. No gate,
query/refresh/balance phase, routing eligibility change, manual-state ownership
inference, scheduler, cooldown inference, payload change or positive enforcement
ACK is implemented. See `resource-catalog-v1.md` for the DTO, conservative file
replacement semantics, supported mode, recovery, migration and rollback limits.

## Isolation and baseline

All Go commands used the task-owned temporary helper `go.ps1`, with:

```text
GOCACHE=<task-scratch>/cache
GOTMPDIR=<task-scratch>/build
GOPROXY=off
GOSUMDB=off
GOTOOLCHAIN=local
```

Local toolchain: `go1.27.0 windows/amd64`. The existing MinGW GCC directory was
added only to the task process PATH and CC; CGO_ENABLED read back as 1. No global
configuration, security, permissions or proxies were changed. No dependency was
downloaded. The scratch root was
`%TEMP%/cpa-spec66-7cefaa4615a74fb2bc38a879c46aa068`.

The coordinator-supplied selected baseline was already green before the task:
`TestV8ExampleLoadsAndRoundTrips`, `TestManagementV8RoutesShareAccessControl`,
`TestAuthClone_EmptyMapsIsolation`. It is baseline evidence, not a new assertion
red/green cycle or full-suite evidence.

## Public-seam red/green evidence

Primary seam: authenticated management HTTP -> real catalog/file persistence,
watcher and auth manager -> temporary synthetic sources. Supporting public module
tests do not substitute for this HTTP seam. Synthetic upstream trap calls were
zero, and secret/path/URL decoys were absent from inventory DTOs.

Command family for the per-behavior cycles below:

```text
go test ./internal/api -run '^<TestName>$' -count=1 -v -timeout=90s
```

For cycles 19-22 this is the exact red and green Go argv, substituting the table
name. Cycles 1-18 were inherited from the prior task continuation. Their retained
logs were re-read to verify genuine assertion FAIL and later PASS for every table
entry; however, stdout did not retain every original shell argv. Some historical
green invocations included additional tests (for example cycles 2 and 4).
The command family is a reproducible selector, not a fabricated byte-for-byte
claim about those unpreserved historical argv. All final tests were rerun in the
exact aggregate commands below. The historical selector receipt limitation
remains explicitly unresolved.

| Cycle | Public test | Observed assertion red | Result |
| --- | --- | --- | --- |
| 1 | `TestResourceInventoryOptInAndAccess` | New route returned 404 instead of 200. | assertion FAIL -> PASS |
| 2 | `TestResourceInventoryCompleteSources` | Opted-in complete inventory still returned 503. | assertion FAIL -> PASS |
| 3 | `TestResourceIdentityRestart` | Restart changed persistent instance identity. | assertion FAIL -> PASS |
| 4 | `TestResourceIdentityKeyReplacement` | Key replacement did not invalidate the old observation. | assertion FAIL -> PASS |
| 5 | `TestResourceIdentityMembershipAndRename` | Member addition did not fence whole-provider observations. | assertion FAIL -> PASS |
| 6 | `TestResourceIdentityUnprovenBindings` | Copied/deleted markers were incorrectly considered ready. | assertion FAIL -> PASS |
| 7 | `TestResourceInventorySingleWriter` | Second catalog writer returned ready inventory. | assertion FAIL -> PASS |
| 8 | `TestResourceInventoryWatcherPublication` | Stale source observation still returned 200. | assertion FAIL -> PASS |
| 9 | `TestResourceInventoryPublicationRecovery` | The original publication-failure fixture lost instance identity (see revised cycle 22). | assertion FAIL -> PASS |
| 10 | `TestResourceInventoryNativeGroupRoundTrip` | Typed native-group save lost issued group/resource bindings. | assertion FAIL -> PASS |
| 11 | `TestResourceInventoryStorageIntegrity` | Corrupt ledger/key still advertised a ready inventory. | assertion FAIL -> PASS |
| 12 | `TestResourceInventoryAuthFileLifecycle` | A genuinely persisted new file source was not published by the watcher hook. | assertion FAIL -> PASS |
| 13 | `TestResourceInventoryUnsupportedStorage` | A non-file manager store was advertised as supported. | assertion FAIL -> PASS |
| 14 | `TestResourceInventoryAmbiguitySurvivesRestart` | Duplicate identity restart returned unavailable instead of complete ambiguous rows. | assertion FAIL -> PASS |
| 15 | `TestResourceInventoryStateDirectoryCannotOverlapAuth` | Catalog state was allowed to overlap and pollute the auth source directory. | assertion FAIL -> PASS |
| 16 | `TestResourceIdentityMarkersReachAuthManager` | Config/file source markers did not reach the auth manager. | assertion FAIL -> PASS |
| 17 | `TestResourceIdentityCopiedEnvelopeFencesObservation` | Byte-identical replacement of a marked physical envelope retained a ready observation. | assertion FAIL -> PASS |
| 18 | `TestResourceInventoryFilteredRowsSurviveTypedSave` | Typed save dropped filtered/keyless configured rows. | assertion FAIL -> PASS |
| 19 | `TestResourceIdentityInheritedCredentialsRoundTrip` | Inherited-header flattening changed credential incarnation. | assertion FAIL -> PASS |
| 20 | `TestResourceInventoryStorageFailureCannotBeHealedByWatcher` | Watcher publication silently healed damaged durable identity state. | assertion FAIL -> PASS |
| 21 | `TestResourceInventoryNestedFileSources` | Recursive file-store source was omitted from complete inventory. | assertion FAIL -> PASS |
| 22 | `TestResourceInventoryPublicationRecovery` | The revised real backup obstruction did not stop publication (old random backup naming). | assertion FAIL -> PASS |

Each listed red was an assertion failure, not "no tests to run" or a compiler
failure. A cycle-6 intermediate implementation attempt had a syntax error; it is
not counted as red/green evidence. The initial cycle-12 fixture did not request
creation of a disabled auth file and was not a valid red; the subsequent retained
`cycle12-red-valid.log` first verified that the synthetic file existed. Cycle 22
replaced the earlier recovery fault injector with a real backup-path obstruction,
leaving the existing ledger intact. The old test could no longer serve as a
publication fault injector after storage-integrity checks were strengthened; it
was not skipped or relaxed.

Additional HTTP coverage (not claimed as new observed assertion-red cycles):
`TestResourceInventoryV8ConfigHTTPRoundTrip` and
`TestResourceInventoryRecoveryRejectsConflictingEdit`. The complete-source test
also asserts manual disabled/excluded/weight counts and separate readiness axes.

Supporting public catalog tests:
`TestResourceIdentityCASAndObservationFences`,
`TestResourceIdentityExplicitRebind`, and
`TestResourceInventoryWriterProcess` (real child process lock contention).
The resumed focused run exposed a real rebind regression after the physical-file
seal change; the exact failing aggregate command was:

```text
go test ./internal/api ./internal/resourcecatalog -run '^(TestResource|TestResourceIdentity|TestResourceInventory)' -count=1 -v -timeout=120s
```

HTTP tests passed; `TestResourceIdentityExplicitRebind` failed with resource catalog
version conflict. After separating content seals from freshness/file seals, this
exact command passed:

```text
go test ./internal/resourcecatalog -run '^TestResourceIdentityExplicitRebind$' -count=1 -v -timeout=90s
```

## Original-worker verification results (historical receipts)

Within the original worker pass, imports were grouped and all changed/new Go
files were formatted with `gofmt -w` before the following runs. Follow-up source
changes and their verification are recorded separately below.

**Focused HTTP/module suite: PASS** (23 HTTP tests plus 3 public module tests):

```text
go test ./internal/api ./internal/resourcecatalog -run '^TestResource' -count=1 -v -timeout=120s
```

**Related package regression: PASS** (15 packages with tests, 2 with no test files):

```text
go test ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/... -count=1 -timeout=180s
```

This covered config, watcher/diff/synthesizer, API/management/middleware, auth store,
SDK lifecycle/auth/registry/executor/session/usage, and catalog tests. The two
"no test files" packages are not additional test passes. This is not `go test ./...`.

**Vet: PASS** (empty diagnostic log):

```text
go vet ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

**Required temporary-output cmd/server compile: emitted an EXE, but not accepted
as a clean verification pass because of an actual permission diagnostic**:

```text
go build -o <task-scratch>/cli-proxy-api-check.exe ./cmd/server
```

The combined command session reported exit 0 and the output binary existed, but
Go attempted a stat-cache write outside the task cache and reported:

```text
go: writing stat cache: open C:\Users\will\go\pkg\mod\cache\download\github.com\router-for-me\!c!l!i!proxy!a!p!i\v8\@v\v8.0.15.info873157875.tmp: Access is denied.
```

The binary was never executed. This diagnostic was treated as the user's explicit
stop boundary, not suppressed, retried with flags, worked around with another
cache/tool, or presented as an unqualified build success. No further implementation
or Go verification was attempted afterward.

**Race: NOT RUN** after the permission stop. Existing GCC availability does not
constitute race evidence. **Linux cross-compile/runtime: NOT RUN. Full repository
suite: NOT RUN.** These are remaining acceptance, not skipped failing tests.

## Remaining acceptance and handoff

A clean owner-approved build environment and race verification remain required
before accepting this slice. The owning coordinator must also review the source,
Linux behavior and operational recovery/rollback/rebind limitations. No production
mode/version, shared filesystem, power-loss recovery or deployment was certified.
Nested auth sources are inventoried, but external nested fsnotify publication is
not added; stale reads fail closed until an existing local write/reload/restart
reconciles them. Same-account token refresh is conservatively fenced because no
trusted continuity proof is available. Automatic downgrade export and an owner
rebind command are not delivered. Gate/routing/refresh/balance/Home/plugin/other
storage phases remain separate work. See the main document for all limitations.

Final inspection uses read-only Git operations with GIT_OPTIONAL_LOCKS=0. The
branch/HEAD remained pinned; that worker reported no staged changes. Accidental CRLF churn in this
task's two tracked documentation/template edits was removed as output cleanup,
not an implementation change or a workaround for the Go permission diagnostic.
Scratch cleanup was attempted only for the exact task-owned directory, using
native PowerShell Remove-Item with an absolute-target guard. The tool rejected the
command as "blocked by policy" before execution. No alternate deletion mechanism,
permission change or retry was attempted. The task-owned scratch/cache/build/log
folder and never-executed EXE therefore remain for the coordinator's approved
cleanup. Input handoff files, shared caches and other directories were not removed.
The original worker reported a clean unstaged `git diff --check` before the
rejected cleanup command; the rejected command made no Git or filesystem change.
That check did not include the then-untracked main document and was NOT a
full-candidate whitespace pass. The coordinator's later immutable staged check
exposed CRLF trailing whitespace throughout `docs/resource-catalog-v1.md`. The
follow-up normalizes only that document to LF without losing its original text;
the immutable staged candidate is deliberately not rewritten to hide this result.

## Changed paths

- `.gitignore`
- `config.example.yaml`
- `docs/resource-catalog-v1-verification.md`
- `docs/resource-catalog-v1.md`
- `internal/api/handlers/management/handler.go`
- `internal/api/handlers/management/resources.go`
- `internal/api/resource_inventory_test.go`
- `internal/api/server.go`
- `internal/api/server_management_v8.go`
- `internal/config/config.go`
- `internal/config/config_types.go`
- `internal/config/config_v8.go`
- `internal/config/resource_binding.go`
- `internal/config/vertex_compat.go`
- `internal/resourcecatalog/catalog.go`
- `internal/resourcecatalog/catalog_test.go`
- `internal/resourcecatalog/durable_other.go`
- `internal/resourcecatalog/durable_unix.go`
- `internal/resourcecatalog/durable_windows.go`
- `internal/resourcecatalog/filestore.go`
- `internal/resourcecatalog/journal.go`
- `internal/resourcecatalog/lock_other.go`
- `internal/resourcecatalog/lock_unix.go`
- `internal/resourcecatalog/lock_windows.go`
- `internal/resourcecatalog/mode.go`
- `internal/resourcecatalog/preconditions.go`
- `internal/resourcecatalog/seals.go`
- `internal/resourcecatalog/source.go`
- `internal/resourcecatalog/sourceid_linux.go`
- `internal/resourcecatalog/sourceid_other.go`
- `internal/resourcecatalog/sourceid_windows.go`
- `internal/watcher/dispatcher.go`
- `internal/watcher/events.go`
- `internal/watcher/synthesizer/config.go`
- `internal/watcher/synthesizer/file.go`
- `internal/watcher/synthesizer/resource.go`
- `internal/watcher/watcher.go`
- `sdk/auth/filestore.go`
- `sdk/cliproxy/auth/classification.go`
- `sdk/cliproxy/auth/conductor_selection.go`
- `sdk/cliproxy/resource/types.go`
- `sdk/cliproxy/service_lifecycle.go`
- `sdk/cliproxy/types.go`
- `sdk/cliproxy/watcher.go`

## Follow-up review corrections (October 25, 2026)

This is local correction/review evidence, not a completed T01/SPEC66/SPEC79,
publication, deployment, model-routing certification, or new permission grant.
No nested agents, route skill, service startup, provider/OAuth calls, browser,
manager changes, index/ref changes, or denied-action workaround were used.

### Immutable candidate and isolation

At entry, branch/HEAD matched the pinned values above and the working tree equaled
the index. The coordinator-staged 44-file binary diff against the base matched
SHA256 `75e0936a791640abc446b658a32dfabfbd1074568523f60aeb5702b98ff33c51`.
Corrections are UNSTAGED on top of that candidate; nothing is committed. A local
expected-file hash guard checks for unrelated/concurrent source changes and for
any change to the immutable staged diff before edits.

All new Go runs use this new task-owned scratch, not the protected old scratch:

```text
C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9
```

Its `go.ps1` sets GOCACHE=`<scratch>/cache`, GOTMPDIR=`<scratch>/build`,
GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local, and process-local PATH/CC for the
already-existing MinGW GCC. GOMODCACHE is unchanged; no dependency cache was
copied, no downloads/network were used, and no permission/security settings were
changed. Toolchain readback was `go1.27.0 windows/amd64`. Full command output is
retained under `<scratch>/logs`; the helper is the executable command prefix for
all exact Go argv below.

### Per-finding verdicts and actual red/green receipts

All five primary hypotheses were reproduced and fixed. Tests drive public module
or management HTTP interfaces with real catalog/config/filesystem state. The
physical-replacement race substitutes only a per-catalog durable filesystem writer
that delegates real writes and replaces a real file at the journal boundary; no
business test flag, fake catalog state machine, timing sleep, or new public hook
is used. The watcher test actually calls Start and observes callback plus runtime
auth synthesis, not direct ReloadConfig calls.

Each command below is prefixed with `& '<scratch>\go.ps1'`. Red and green used
the exact same listed Go argv:

| Finding | Exact argv | Observed assertion red -> green | Logs |
| --- | --- | --- | --- |
| 3 | `test ./internal/api -run '^TestResourceInventoryFirstInitializationRepair$' -count=1 -v -timeout=90s` | Repaired first initialization still returned HTTP 503 -> HTTP inventory succeeds and survives another reopen. | `03-red.log`, `03-green.log` |
| 5 | `test ./internal/api -run '^TestResourceInventoryManualScalarSpellings$' -count=1 -v -timeout=90s` | Accepted literal YAML disabled/weight spellings were misreported -> all seven literal fixtures pass, including inherited flags. | `05-red.log`, `05-green.log` |
| 5, inheritance | `test ./internal/api -run '^TestResourceInventoryManualInheritanceOverrides$' -count=1 -v -timeout=90s` | Explicit empty child exclusions were reported as inherited true -> `[true, false, true]` for inherited / empty override / null inheritance. | `05-inheritance-red-valid.log`, `05-inheritance-green.log` |
| 1 | `test ./internal/resourcecatalog -run '^TestResourceIdentityPublicationRejectsPhysicalReplacement$' -count=1 -v -timeout=90s` | Real byte-identical auth replacement after source collection retained a ready observation target -> publication/Check/recovery never accept that old target. | `01-red.log`, `01-green.log` |
| 4 | `test ./internal/api -run '^TestResourceInventoryAmbiguousGroupsAllowManualSave$' -count=1 -v -timeout=90s` | Typed exclusion save failed with `ambiguous resource group` -> save/reorder succeeds without dropping copied groups, markers, keyless rows or the empty group; duplicate identities stay ambiguous across restart. | `04-red-valid.log`, `04-green.log` |
| 2 | `test ./internal/watcher -run '^TestResourceInventoryStartedWatcherSurvivesReplacements$' -count=1 -v -timeout=90s` | On this Windows host, a started watcher migrated a new key, then stopped delivering the next manual disable -> owned replacement, disablement, external replacement, and another disable all reach callback/runtime state. | `02-windows-baseline.log`, `02-green.log` |

The finding-1 recovery-only subcase was already conservative in the initial
module test; its publication subcase supplied the actual red. Additional HTTP
recovery evidence now verifies the stronger requirement: a real partial
catalog-owned source replacement recovers without rebinding existing credentials,
whereas a byte-identical copy of that replacement is refused. Signed format-2
intents pin all source file seals and record an owned prepared-file receipt before
rename. Content equality alone cannot authorize a replacement.

First initialization now saves an empty issued ledger before fallible source
migration. Supplemental authenticated HTTP tests still refuse deletion of an
established catalog.json or seal.key without recreating it or rewriting sources.

Duplicate native groups retain transient source-document-bound manual-save
provenance through flattening. It is stripped on write, excluded from JSON, and
never grants durable identity validity. Unknown/stale ambiguous provenance is not
rebound by name/key/position. The catalog continues to block copied resource IDs.

Literal auth fixtures additionally cover JSON 0.0, 0e0, -0.0, string "0", positive
weight, non-boolean disabled metadata, and invalid fractional/string weights.
Runtime-parser-invalid YAML manual values remain unavailable rather than silently
coerced. No runtime parser semantics were changed.

### Setup/fixture failures, not counted as assertion-red evidence

- `05-inheritance-red.log`: the first Codex inheritance fixture omitted its
  required base URL and therefore did not establish the intended runtime state.
  It was corrected before the valid assertion red above.
- `04-red.log`: the first fixture assumed Codex filters keyless rows. It does not.
  The valid fixture uses Meta's actual runtime keyless-row filter and then fails
  at the intended manual-save assertion.
- `01-03-safeguards.log`: the first HTTP partial-recovery fixture used an unsuitable
  `os.SameFile` check for its replacement precondition on Windows. The corrected
  fixture verifies newly persisted owned source markers before the obstruction;
  `01-03-safeguards-valid.log` records all four safeguards passing. The independent
  module race fixture separately checks real physical file IDs before/after copy.

No setup failure, missing selector, or no-test package is claimed as a passing
regression or assertion red. Existing tests were not disabled or weakened.

### Standards and documentation corrections

Unexpected temporary-file removal and failed-initialization lock-close errors now
emit constant redacted logrus warnings with no path, credential, or raw error text.
Expected not-exist after a successful rename is ignored. This is new catalog code,
not an attempt to delete the protected old scratch.

The duplicated config-plus-recursive-auth traversal now has one small collection
helper. Physical-plus-content and content-only projections remain explicit, with
their previous source coverage and error handling retained.

The immutable staged `git diff --cached --check` was run and returned exit 2 for
CRLF trailing whitespace in the main new document (`staged-whitespace.log`). Only
that document's CRLF line endings were normalized, with a byte-level check that
normalization itself lost no text. No Git attributes/config or unrelated file
formatting was changed. The staged check necessarily remains failing while the
candidate is immutable; the corrected working-tree document is checked separately.

### Follow-up final verification

The focused regression command passed after the final Go changes:

```text
test ./internal/api ./internal/resourcecatalog ./internal/watcher -run '^TestResource' -count=1 -v -timeout=120s
```

All three packages passed (`focused.log`). Supplemental resource tests also cover
partial owned-publication recovery, refusal of a copied replacement on recovery,
established ledger/key loss, invalid manual scalars, and auth JSON scalar forms.

#### Related packages: first run failed, exact rerun passed

```text
test ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/... -count=1 -timeout=180s
```

The first run (`related.log`, exit 1) had 14 passing packages, one failing package
(`sdk/cliproxy`), and two packages with no test files. All three failures are
retained, not overwritten by the later successful rerun:

- `TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch`,
  `service_auth_sync_test.go:763`: auth A hook wait blocked on unrelated auth B
  registration.
- `TestEndToEndAuthFileReplacement_RestoresModelsInV1ModelsWithoutRestart`,
  `service_auth_sync_test.go:1397`: expected `gpt-6-astra` absent after terminal auth
  failure, but it remained present.
- `TestEndToEndAuthFilePatch_RestoresModelsInV1ModelsWithoutRestart`,
  `service_auth_sync_test.go:1557`: the same absent-model expectation failed.

The first SDK test passed in isolation (`related-sdk-focused.log`, exit 0):

```text
test ./sdk/cliproxy -run '^TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch$' -count=1 -v -timeout=90s
```

One exact rerun of the full related command passed: 15 packages passed, two had no
test files (`related-rerun.log`, exit 0). No code/test changes or selector/timeout
changes were made between these related runs. This is not proof of a baseline
root cause or a clean reliability gate; the first-run SDK failures still require
coordinator assessment. No unrelated SDK repair or weakened assertion was added.

#### Vet: passed

```text
vet ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

Exit 0 with no diagnostics. A confirming run has a deliberately created empty
`vet-confirmed.log` and an exit-code/argv receipt in `vet-confirmed.json`; an earlier
no-output Tee pipeline did not create the initially expected `vet.log`. Its absence
was a receipt-reading setup issue, not an assertion red or a permission failure.

#### Race: corrected packages passed; unchanged middleware failed

```text
test -race ./internal/resourcecatalog ./internal/config ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/auth -count=1 -timeout=180s
```

The scoped race run failed (`race.log`, exit 1): nine packages passed, including all
follow-up-modified packages, and `internal/api/middleware` failed. The race was
reproduced in isolation (`race-middleware-focused.log`, exit 1):

```text
test -race ./internal/api/middleware -run '^TestRequestLoggingMiddleware_StreamingPreservesFullUUIDForLogger$' -count=1 -v -timeout=90s
```

`ResponseWriterWrapper.processStreamingChunks` reads `chunkChannel` at
`response_writer.go:257` while `Finalize` writes it at line 309. The source and test
paths `response_writer.go`, `request_logging.go`, and `request_logging_test.go`
are unchanged from the pinned HEAD. This is concrete unchanged-code race evidence,
not a broad clean-race claim. Repair is outside this bounded follow-up. The complete
SDK service package was not included in the race run.

#### Formatting, whitespace, and immutable candidate

- `gofmt -w` was applied to the 11 follow-up Go files; the final formatting pass
  changed no bytes, and `gofmt -l` returned no paths.
- `git diff --check --` and
  `git diff --check a4acc9f752bd46571f737a10c04bf413656ab06b --` passed for tracked
  working-tree corrections/the corrected full tracked candidate.
- Each of the four new untracked tests was checked separately with
  `git diff --no-index --check -- <scratch>/empty <file>`: no whitespace diagnostics.
  An initial audit incorrectly required exit 0; Git returns exit 1 for differing
  files even with no check diagnostics. The corrected audit accepts only 0/1 with
  empty output. That audit-script error was not a regression red or a whitespace
  defect. The receipt is `final-whitespace.json`.
- Both documents are LF-only. The staged candidate still intentionally fails
  `git diff --cached --check` for its original main-document CRLF whitespace; no
  index change was used to conceal or repair the immutable review snapshot.
- Final branch/HEAD match the pinned values. The 44-file staged binary diff still
  has SHA256 `75e0936a791640abc446b658a32dfabfbd1074568523f60aeb5702b98ff33c51`.
  All follow-up corrections remain UNSTAGED, including the four untracked tests.
  The expected-file hash guard detected no unrelated/concurrent source changes.

### Exact follow-up changed paths

These nine existing candidate paths have additional unstaged corrections:

- `docs/resource-catalog-v1-verification.md`
- `docs/resource-catalog-v1.md`
- `internal/config/config_v8.go`
- `internal/config/resource_binding.go`
- `internal/resourcecatalog/catalog.go`
- `internal/resourcecatalog/filestore.go`
- `internal/resourcecatalog/journal.go`
- `internal/resourcecatalog/source.go`
- `internal/watcher/events.go`

These four test paths are new, untracked, and unstaged:

- `internal/api/resource_inventory_followup_test.go`
- `internal/resourcecatalog/publication_test.go`
- `internal/resourcecatalog/source_validation_test.go`
- `internal/watcher/resource_catalog_test.go`

### Remaining gates and scratch disposition

- The denied clean `cmd/server` build remains **BLOCKED**, not retried. Its prior
  shared module stat-cache denial is not suppressed, rerouted to another context,
  or bypassed by changing GOMODCACHE/security/proxy/flags. New package-test logs
  contain no shared stat-cache write or permission-denied diagnostics.
- **Linux runtime is NOT verified.** The real started-watcher Windows red/green
  does not establish Linux runtime/durability correctness. No Linux runtime or
  cross-compile evidence is fabricated here.
- The first-run SDK test failures and reproducible unchanged middleware race remain
  open assessment/repair gates. The full-repository suite, clean server build,
  actual power-loss recovery, deployment/rollback, and independent coordinator
  acceptance are not complete.
- The protected original scratch was not removed, modified, reused, or retried.
  The NEW follow-up scratch is retained for coordinator review of exact logs and
  receipts; no cleanup was attempted. This is retained local evidence, not a
  claim of cleanup completion.
- The bounded five-finding correction slice is delivered for review, not full
  T01/SPEC66/SPEC79, routing/gate/provider-phase, deployment, or publication
  completion. No schedules, task cards, public comments, or new permissions.


## Pass 4: preservation repair

Dates: October 25-26, 2026, Asia/Singapore. This is an owning-project local repair,
not publication, deployment, a permissions exception, or full SPEC acceptance.
No agents were spawned and no other model/task was delegated. Only this source
workspace was edited; all corrections remain **UNSTAGED** for coordinator freeze.

### Frozen input, independent review and blocked comparison

The coordinator had already staged all prior corrections and four follow-up test
files into **48 files, 4,906 additions / 42 deletions**. At entry, the working tree
equaled the index with no untracked files. Branch and HEAD/base were:

```text
claude/spec66-resource-catalog
a4acc9f752bd46571f737a10c04bf413656ab06b
```

The exact SHA256 of `git diff --cached --binary <base> --` was and remains:

```text
960e7f9fb56db4774f5864f2723fe1aed0d56b5bc04c292f67b7b406c473e5db
```

There is no new commit; `base...HEAD` would be empty and is not the candidate.
Before each edit, read-only Git checks and expected content hashes verified the
index, branch, HEAD, allowed unstaged paths and each prior correction. No index,
ref, branch, staging, commit, fetch, push or publication action was performed.

The supplied latest two-axis review reported **Standards: 0 hard violations / 0
smells; Spec: three new static preservation findings**. That is review of the
frozen 48-file snapshot, not review or execution proof of these unstaged repairs.
The original five findings were already corrected and are not reclassified.

Before this repair, the coordinator independently ran `^TestResource` on the
48-file candidate: internal/api 17.641s, internal/resourcecatalog 6.416s,
internal/watcher 4.878s. Those are current-candidate resource results, not a full
suite, baseline comparison, or evidence that the three new findings were fixed.

The separately attempted baseline comparison executed **ZERO comparison tests**:
writing `frozen-before.json` raised PermissionError in
`C:\Users\will\AppData\Local\Temp\cpa-spec66-baseline-verify-20261025-gbvujhm7`,
and `git archive HEAD` exited 128 because object
`5cb02483dd3054a0957cc65195f9d28f54cba990` was missing with lazy fetching disabled.
That task/directory/materialization was not retried, relocated or duplicated here;
no objects were fetched and no alternative baseline/cache was prepared. Ordinary
current-candidate package tests below are distinct from that blocked comparison.

### Commands and actual red/green receipts

All Go argv below have this exact executable/owner prefix:

```powershell
& 'C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9\go.ps1' -C 'D:\WILL\AGENT\CPA\CLIProxyAPI-source'
```

The existing helper was read before use. Its existing task-local GOCACHE/GOTMPDIR,
GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local and process-local existing GCC were
reused. GOMODCACHE, permissions, settings and hooks were unchanged. No new cache
root or dependency download was prepared. Evidence is the task's command/stdout
transcript; no new scratch log file was needed. All fixture files, keys and
upstreams were temporary and synthetic. No normal server/provider/OAuth service,
8317, real config/environment/database, browser or yosef-server was accessed.

Each row's red and green used the same listed argv, with a targeted code change
between runs, not retry-until-green:

| Finding / behavior | Exact argv after the prefix | Actual assertion red -> green |
| --- | --- | --- |
| P1 compatibility provider association | `test ./internal/api -run '^TestResourceInventoryCompatibilityFilteredProviderSave$' -count=1 -v -timeout=90s` | Exit 1, api 4.219s: missing-base B disappeared, synthesized auths=2 instead of 1, and controlled requests used A=2/B=2 rather than A=4/B=0. Disabled B's source data/inventory was changed. -> Exit 0, api 4.468s: both original cases preserve B and route four requests through A only. |
| P2 represented child association | `test ./internal/api -run '^TestResourceInventoryFilteredMemberProjection$' -count=1 -v -timeout=90s` | Exit 1, api 4.944s: across-group collision lost a provider; across/within-group collisions lost the keyless child; missing IDs lost filtered and empty groups. -> Exit 0, api 4.874s: all three original shapes preserve those entries, edit/reorder live rows and honor explicit represented deletion. |
| P3 current-member null inheritance | `test ./internal/api -run '^TestResourceInventoryFilteredNullInheritance$' -count=1 -v -timeout=90s` | Exit 1, api 8.610s: Meta and Vertex unique/duplicate cases changed effective fields and credential exclusions (2 vs 3; duplicate counts also deficient). -> Exit 0, api 8.179s: all 14 family/group cases retain current-member effective values, including explicit empty overrides. |
| P3 shared group defaults | `test ./internal/api -run '^TestResourceInventoryFilteredNullInheritance$/^meta$' -count=1 -v -timeout=90s` | Exit 1, api 4.351s: a weight-only edit erased shared defaults (all nine fields for unique Meta; exclusions/headers/error rules for duplicated Meta). -> Exit 0, api 4.161s after retaining group defaults and encoding explicit empty/zero child overrides. Final tests assert all nine shared fields for all seven native families. |
| Provenance namespace safeguard | `test ./internal/api -run '^TestResourceInventory(ManualProjectionRefusal\|CompatibilityProjectionEdits)$' -count=1 -v -timeout=90s` | Exit 1, api 6.801s: a child reference used as a provider reference was accepted when its child list was removed; positive edit cases passed. -> Exit 0, api 6.711s after separating known provider/child reference sets. In this table the escaped pipe denotes the literal regex alternative `|`, not a shell pipeline. |
| P2 stale represented deletion safeguard | `test ./internal/api -run '^TestResourceInventoryManualProjectionRefusal$' -count=1 -v -timeout=90s` | Exit 1, api 6.774s: Meta/compatibility stale delete-all and deleted-sibling views were accepted despite a concurrent synthetic source addition. -> Exit 0, api 6.273s after the private source-document fence; refused saves preserve exact source bytes. |

P1's disabled-only fixture also established an important distinction: the parser
retains a disabled compatibility provider and the synthesizer filters it. With A
first and disabled B last, the original red copied A into B rather than executing
B. The actual A=2/B=2 red was the missing-base case. Final coverage includes both
source orders for both conditions and independently checks synthesis, manager
contents, persisted B data and real controlled executor requests. Inventory
counts are not used as a substitute for routing evidence.

The deletion fence is a private string inside the scalar-only catalog config,
not a new public field/API or persisted identity. It is checked before a typed
save can infer deletion from missing entries, including when no surviving child
reference remains. Public CloneForRuntime plus fresh delete-all success verifies
that legitimate copied/manual views are not blanket-blocked. Reload the source
before reusing a stale projection. This is not a new multi-writer transaction or
an atomic cross-file guarantee.

### Retained fixture/intermediate failures

- The first P1 run (exit 1, api 4.650s) also used the wrong internal provider name
  and reached `auth_not_found: no auth available`. Source-preservation assertions
  failed too, but this was not counted as the valid routing red. Correcting only
  the fixture provider name produced the valid A=2/B=2 red before any P1 fix.
- The first P2 run (exit 1, api 5.405s) had genuine collision reds, but its
  missing-ID branch reopened the catalog, which issued fresh markers. That branch
  was not evidence for missing IDs. The corrected supporting typed-save fixture
  keeps the source unmarked until after saving; the next run supplied all three
  valid assertion reds.
- The first broad inheritance draft (exit 1, api 8.233s) also counted provider-row
  exclusion flags and failed in all 14 cases. A narrower current-credential oracle
  isolated the null bug. Group-default loss was not dismissed: the later explicit
  source-field red above led to preserving those defaults as well. Existing tests
  were not weakened or removed; final assertions cover current credential values,
  all shared group fields and explicit clear behavior independently.
- An intermediate full related run (exit 1) found three genuine regressions caused
  by attaching projection metadata to ordinary unmarked opt-out configs:
  `TestV8ExampleLoadsAndRoundTrips`, `TestV8MigrationPreservesLegacySemantics`, and
  `TestV8AliasesAndMergeKeys` (config 3.747s). The production change was restricted
  to opted-in or catalog-marked sources. The same unchanged suite then passed,
  and the final full config run below also passed. No test or timeout was changed
  to mask this regression.
- The stale-deletion safeguard first reproduced with a changed source name
  (exit 1, api 6.864s), then with an actual concurrent synthetic key addition
  (exit 1, api 6.774s) before its fix.
- A proposed Node read-only subprocess guard returned `spawnSync git EPERM` before
  any Node edit/test/log write. That checker was abandoned; no write or test was
  retried there. The already-working owner-shell read-only checks were used for
  ordinary source editing. No actual edit/test/log permission denial occurred.

### Final coverage matrix and verification

Seven new public-seam tests contain **44 subcases**:

| Test suffix after `TestResourceInventory` | Cases | Coverage |
| --- | ---: | --- |
| CompatibilityFilteredProviderSave | 4 | Disabled/missing-base provider, each source order, exact B source preservation, duplicate ambiguity, runtime/manager eligibility, controlled A-only requests, no JSON/YAML provenance leak. |
| FilteredMemberProjection | 5 | Cross-group and within-group duplicate child IDs; missing all/child/group IDs; empty groups; live key/weight edits and reorder; positive represented deletion. |
| FilteredNullInheritance | 14 | Gemini, Interactions, Vertex, Codex, Claude, xAI, Meta; unique/duplicated groups; omitted/null inheritance versus explicit empty arrays/maps; all nine shared group defaults and parser-effective row values. |
| ManualProjectionRefusal | 14 | Native/compatibility unknown, missing and stale projection; deleted-all/deleted-sibling concurrent additions; compatibility child/provider namespace and parent association; byte-identical no-write refusal. |
| CompatibilityProjectionEdits | 3 | Unique/duplicate/missing IDs; provider rename and endpoint/header/key edits, reorder, represented provider/child deletion; exact filtered/disabled/empty sources; public synthesis eligibility. |
| ManualProjectionDeleteAll | 2 | Fresh cloned native/compatibility views may delete every represented row while filtered and empty sources remain unroutable. |
| InheritedManualClears | 2 | Unique/duplicate Meta; explicit nullable/scalar/map/list clears do not re-inherit old settings or change represented/filtered siblings; shared exclusion default remains. |

Final focused argv (exit 0, api **16.040s**):

```text
test ./internal/api -run '^TestResourceInventory(CompatibilityFilteredProviderSave|FilteredMemberProjection|FilteredNullInheritance|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)$' -count=1 -timeout=90s
```

Final complete resource selector (exit 0; api **26.617s**, resourcecatalog
**2.541s**, watcher **2.485s**), including the original five correction tests:

```text
test ./internal/api ./internal/resourcecatalog ./internal/watcher -run '^TestResource' -count=1 -timeout=180s
```

Final full related config/catalog/watcher packages (exit 0; config **3.247s**,
resourcecatalog **3.661s**, watcher **5.189s**, watcher/diff **2.371s**,
watcher/synthesizer **2.678s**):

```text
test ./internal/config ./internal/resourcecatalog ./internal/watcher/... -count=1 -timeout=180s
```

Final vet (exit 0, no diagnostics):

```text
vet ./internal/config ./internal/resourcecatalog ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

Final focused changed-package race passed (exit 0; config **10.790s**, api
**122.609s**):

```text
test -race ./internal/config ./internal/api -run '^(TestV8|TestResourceInventory(CompatibilityFilteredProviderSave|FilteredMemberProjection|FilteredNullInheritance|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)$)' -count=1 -timeout=180s
```

An earlier intermediate-code race run of `^(TestResource|TestV8)` passed config
10.630s / api 179.407s. It is retained as intermediate evidence, not substituted
for final focused-race readback and not a full clean-race claim.

Final formatting and read-only whitespace checks passed: `gofmt -w` changed no
bytes in the five tested Go files; `gofmt -l` returned no paths; both docs are
LF-only. `git diff --check --`, `git diff --cached --check`, and
`git diff --check a4acc9f752bd46571f737a10c04bf413656ab06b --` all exited 0.
Unlike the historical 44-file snapshot, the coordinator's current 48-file staged
snapshot has no whitespace diagnostic. Branch, HEAD and the canonical staged-diff
SHA256 remain unchanged. Only these seven correction paths are unstaged, with no
untracked files or unexpected/concurrent source changes:

- `internal/config/config.go`
- `internal/config/config_v8.go`
- `internal/config/config_yaml.go`
- `internal/config/resource_binding.go`
- `internal/api/resource_inventory_followup_test.go`
- `docs/resource-catalog-v1.md`
- `docs/resource-catalog-v1-verification.md`

### Gates deliberately left open

- The prior clean `cmd/server` build remains **BLOCKED** by the actual shared
  module stat-cache write denial at
  `C:\Users\will\go\pkg\mod\cache\download\github.com\router-for-me\!c!l!i!proxy!a!p!i\v8\@v\v8.0.15.info873157875.tmp`.
  No build/run retry, alternative GOMODCACHE/flags/cache/security context or peer
  was used. Passing package tests/vet/race do not turn that gate green.
- The three earlier SDK failures and reproducible unchanged middleware race remain
  unresolved. Neither was reclassified by these scoped passes. The final race
  command does not include middleware or the complete SDK service package.
- No baseline/candidate comparison tests ran. No full-repository/full-SPEC pass,
  Linux runtime/durability, real power-loss, deployment/rollback, operational
  clone/rebind, gate/query/routing/balance/hourly/provider-phase acceptance or
  independent review of these unstaged corrections is claimed.
- The old protected scratch and retained prior follow-up evidence were not
  removed, reused or cleaned. No cleanup attempt was made in this pass.


## Pass 5: owned saves, endpoint clears and effective credential seals

Date: **October 26, 2026** (Asia/Singapore). Local owner-project repair only;
corrections remain **unstaged and uncommitted**. This is not full T01/SPEC66/70/79
acceptance, a deployment, or an independent review of the new corrections.

### Frozen input and independent prior evidence

- Branch `claude/spec66-resource-catalog`; HEAD/base
  `a4acc9f752bd46571f737a10c04bf413656ab06b`.
- Current immutable input: **49 staged files, +5911/-45**, working tree equal to
  the index with no untracked files at entry. The authoritative comparison was
  `git diff --cached <base> --`, not the empty commit-range diff.
- Canonical `git diff --cached --binary <base> --` SHA256 at entry and intermediate
  readbacks: `6cdcdf1c535855adc46c462adffc11c9ecfe624683dd0b9a27b0920c48b6a83e`.
- The supplied October 26 Pass 4 Standards report found **0 hard violations / 0
  actionable heuristic smells** in that staged input. The separate static Spec
  report found the **three new P2 findings** repaired below. Static review did not
  reproduce them. The original five and next three preservation corrections are
  not reclassified or removed.
- Independently supplied checks on the frozen input, before these repairs:
  - `test ./internal/config ./internal/api ./internal/resourcecatalog ./internal/watcher/... -count=1 -timeout=180s`:
    exit 0; config 4.868s, api 64.402s, catalog 42.114s, watcher 44.267s,
    watcher/diff 41.565s, watcher/synthesizer 41.785s.
  - `vet ./internal/config ./internal/resourcecatalog ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...`:
    exit 0, empty diagnostics.
  - The supplied focused `-race` config/api selector covering `TestV8` and the
    seven Pass 4 preservation tests: exit 0; config 10.423s, api 127.157s.
    This was not full-repository, middleware or full-SDK race evidence.
  These independent checks do **not** prove the three new findings fixed.

### Execution boundary and actual TDD receipts

Every Go argv below used this exact approved prefix, with `-C` before the
subcommand:

```powershell
& 'C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9\go.ps1' -C 'D:\WILL\AGENT\CPA\CLIProxyAPI-source'
```

The existing helper was read and reused unchanged: task-local GOCACHE/GOTMPDIR,
GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local, existing process-local GCC, unchanged
GOMODCACHE. No new cache/baseline root, download, network/provider/OAuth/browser,
8317, normal service startup or yosef-server access was used. Sources, keys,
state and upstreams were synthetic/temporary; receipts are command stdout, not a
new evidence scratch tree. No nested agents, model substitution, permission
exception, Git mutation, external publication or deployment was performed.

The selected seams were the agreed real management HTTP handlers, public typed
save/parse methods, real catalog reconciliation and HTTP inventory, real watcher
synthesis/auth manager/executor, temp persistence, and a controlled HTTP upstream.
The pre-existing manual `/v0` routes were exercised, not feature-modified. No
business hook, private-method assertion or state-machine mock was introduced.

Each row used the **same argv** for its genuine assertion-red and green, separated
by its minimal production change. Historical receipts above were not rewritten.

| Finding / observed behavior | Exact argv after the prefix | Assertion red -> same-test green |
| --- | --- | --- |
| P2 repeated owned saves / retained manual handler config | `test ./internal/api -run '^TestResourceInventory(OwnedSaveContinuity\|ManualSaveContinuity)$' -count=1 -timeout=90s` | Exit 1, api **4.721s**: both native/compatibility second typed saves rejected their own prior write as stale; both second manual HTTP edits returned 500. -> Exit 0, api **4.579s** after recording emitted-row projections and publishing metadata only after successful persistence. |
| P2 native endpoint clear | `test ./internal/api -run '^TestResourceInventoryNativeEndpointClear$' -count=1 -timeout=90s` | Exit 1, api **7.582s**: Gemini, Interactions and Vertex, each with unique/duplicate IDs, reloaded the previous edited endpoint instead of `""` (six failing subcases). Codex/Claude/xAI/Meta non-omitempty cases did not reproduce that omission bug. -> Exit 0, api **7.749s**, all 14 subcases, after removing the copied group endpoint when the typed endpoint is omitted. |
| Required interaction: an owned edit becomes filtered | `test ./internal/api -run '^TestResourceInventoryOwnedSaveFilteredTransition$' -count=1 -timeout=90s` | Exit 1, api **4.037s**: repeating a Codex clear duplicated its member (2 instead of 1); compatibility duplicated its provider (3 instead of 2). -> Exit 0, api **4.144s** after counting exact, validated input provenance as represented rather than copying it again as untouched filtered data. Runtime still selects only synthetic-B. This is the closely related interaction exposed by enabling successive saves, not new routing behavior. |
| P2 null-inherited credential seal | `test ./internal/api -run '^TestResourceInventoryEffectiveCredentialContinuity$' -count=1 -timeout=90s` | Exit 1, api **4.316s**: Gemini `headers: null` plus inherited group headers changed incarnation and identity/credential versions **1 -> 2** on a weight-only save; omitted and explicit-empty cases passed. -> Exit 0, api **4.266s** after honoring null inheritance. Initial issuance and before/after HTTP targets are asserted independently, not by computing an implementation seal. |
| Same seal finding: canonical effective defaults | `test ./internal/api -run '^TestResourceInventoryEffectiveCredentialDefaults$' -count=1 -timeout=90s` | Exit 1, api **5.109s**: null group headers, null endpoint, Meta's parser default and normalized headers each falsely rotated identity/credential versions **1 -> 2**; explicit empty group headers passed. -> Exit 0, api **4.898s** after typed canonical header/endpoint projection. Physical/source seals and conservative JSON credential sealing were not changed. |

Supporting first-finding/preservation run (exit 0; api **9.972s**, config
**1.269s**):

```text
test ./internal/api ./internal/config -run '^(TestResourceInventory(OwnedSaveContinuity|ManualSaveContinuity|OwnedSaveProjectionLifecycle|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)|TestV8MigrationUpdatesScopeAfterSaving)$' -count=1 -timeout=90s
```

The final seal-support run expanded null/omitted/empty inheritance to **all seven
native families**, and passed real manual credential/config changes, native group
endpoint-conflict refusal and synthetic write-failure cases (exit 0; api
**11.956s**):

```text
test ./internal/api -run '^TestResourceInventory(EffectiveCredentialContinuity|EffectiveCredentialDefaults|NativeEndpointConflict|OwnedSaveWriteFailure|ManualCredentialChanges)$' -count=1 -timeout=90s
```

The ten added test functions contain **52 actual leaf cases**, including 21
native inheritance cases and 14 endpoint family/identity cases. The write-failure
cases use only a synthetic temp file's read-only attribute, restore that fixture
attribute, and prove through public saves that the failed write adopted no false
basis. They do not change ACLs or bypass an environment permission denial.
Fresh clones save successfully; older independent clones and later external
additions/delete-all views are rejected without changing source bytes. Runtime-only
Home/steering values and unrelated caller settings are preserved.

The manual credential test executes five actual HTTP edits without replacing the
handler config: weight, headers, endpoint, key and priority. After every edit it
checks public catalog targets, reparses temp disk data, synthesizes real auths,
uses the real manager/Gemini executor and observes the controlled upstream's path,
header and key. Literal credential/identity versions are **1,2,3,4,4**, config
versions **2,3,4,5,6**; incarnation changes only for real credential edits. An
unchanged restart preserves the final target. No request is sent to a default
provider endpoint.

### Additional failed action, not test evidence

A later combined command to append an **extra** concurrent-manual regression,
format, and run an aggregate focused selector was **rejected before process
creation: blocked by policy**. It executed zero tests and made no source change;
`TestResourceInventoryManualConcurrentSave` was not created. That action/addition
and selector were not retried by another tool, path, flags, context or peer.
The distinct, already requested full-package/vet/focused-race gates below use the
unchanged approved helper, not an alternative execution environment. The separate
extra concurrent-manual case remains unexecuted, not a claimed pass.

Read-only discovery also encountered Windows `rg` path-glob/missing-file errors;
these were inspection/setup mistakes, not permission exceptions or assertion-red
receipts. Existing paths were subsequently inspected. All actual assertion reds
and the policy-rejected action are retained above; there was no retry-until-green
or weakening of old tests.

### Final package gates

Unfiltered complete related packages, exit **0**:

```text
test ./internal/config ./internal/api ./internal/resourcecatalog ./internal/watcher/... -count=1 -timeout=180s
```

- config **3.716s**; api **40.557s**; resourcecatalog **3.340s**;
  watcher **5.055s**; watcher/diff **2.292s**;
  watcher/synthesizer **2.619s**.
- Includes the original five corrections, the next three preservation corrections,
  all ten new tests, existing config migration/clone behavior and related watcher
  regressions. This is not full-repository or full-SDK acceptance.

Vet, exit **0**, empty diagnostics:

```text
vet ./internal/config ./internal/resourcecatalog ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

Focused changed-package race, exit **0**:

```text
test -race ./internal/config ./internal/api ./internal/resourcecatalog -run '^(TestV8|TestResourceInventory(OwnedSaveContinuity|ManualSaveContinuity|OwnedSaveProjectionLifecycle|NativeEndpointClear|OwnedSaveFilteredTransition|EffectiveCredentialContinuity|EffectiveCredentialDefaults|NativeEndpointConflict|OwnedSaveWriteFailure|ManualCredentialChanges|CompatibilityFilteredProviderSave|FilteredMemberProjection|FilteredNullInheritance|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)$|TestResourceIdentity(CASAndObservationFences|ExplicitRebind)$)' -count=1 -timeout=360s
```

- config **10.712s**, api **260.621s**, resourcecatalog **2.368s**.
- This covers the ten new regression functions, seven prior preservation
  functions, config `TestV8*`, and public catalog CAS/rebind regressions. It is a
  **focused** race gate, not all tests in these packages, middleware race, full
  SDK race or a full-repository clean-race claim. The separate extra concurrent
  HTTP regression rejected before creation remains unexecuted.
- No final test/race/vet command failed or was rerun without a code change. The
  genuine targeted red/green cycles and rejected action above are separate
  receipts, not overwritten by these passing final gates.

### Final formatting, freeze and changed paths

All five changed Go files were formatted with `gofmt -w` before their green runs;
read-only final `gofmt -l` returned no paths. Both docs and all five Go files are
UTF-8 without BOM, LF-only. Whitespace and frozen-input readbacks use ordinary
owner-shell read-only Git, with optional index locks disabled. The unchanged
canonical staged binary diff remains:

```text
6cdcdf1c535855adc46c462adffc11c9ecfe624683dd0b9a27b0920c48b6a83e
```

The input is still 49 staged files (+5911/-45). Branch and HEAD/base are unchanged;
no index/ref/branch/staging/commit/fetch/push operation was requested or performed.
Only these seven repair paths differ from the frozen index; no untracked files:

- `internal/config/config_yaml.go`
- `internal/config/config_v8.go`
- `internal/config/resource_binding.go`
- `internal/resourcecatalog/seals.go`
- `internal/api/resource_inventory_followup_test.go`
- `docs/resource-catalog-v1.md`
- `docs/resource-catalog-v1-verification.md`

The original preservation tests were neither edited nor weakened. New assertions
were appended to the existing follow-up test file. No business test hook, global
writer registry/serializer, per-member endpoint model or trusted-token continuity
mechanism was added. The existing save is still caller-serialized; its document
fence is not a new atomic transaction with arbitrary uncooperative source writers.

### Remaining gates and retained failures

- **Clean `cmd/server` build/run remains blocked and was not attempted.** The
  original shared module stat-cache denial at
  `C:\Users\will\go\pkg\mod\cache\download\github.com\router-for-me\!c!l!i!proxy!a!p!i\v8\@v\v8.0.15.info873157875.tmp`
  is not resolved by package-test compilation. No alternate flags, GOMODCACHE,
  security/cache/context/peer or permission exception was used.
- The protected old scratch cleanup was not touched/reused/deleted. Other retained
  evidence was not cleaned. The blocked baseline task still ran **ZERO tests**:
  `frozen-before.json` PermissionError in
  `C:\Users\will\AppData\Local\Temp\cpa-spec66-baseline-verify-20261025-gbvujhm7`,
  plus archive exit 128 for missing object
  `5cb02483dd3054a0957cc65195f9d28f54cba990` with lazy fetch disabled. No retry,
  relocated write, materialization, alternative baseline/cache or fetch occurred.
- The abandoned Node `spawnSync` Git EPERM checker was not retried. The new
  combined-action policy denial and its exact unexecuted case are retained above.
- The historical SDK failures in
  `TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch`,
  `TestEndToEndAuthFileReplacement_RestoresModelsInV1ModelsWithoutRestart`, and
  `TestEndToEndAuthFilePatch_RestoresModelsInV1ModelsWithoutRestart` remain open.
  This pass did not rerun the full SDK test package or infer a baseline cause.
- The unchanged middleware race in
  `TestRequestLoggingMiddleware_StreamingPreservesFullUUIDForLogger` remains open
  and outside this race selector. No unrelated SDK/middleware repair was made.
- No new independent Standards/Spec review of these unstaged changes, baseline
  comparison, full-repository suite, Linux runtime/durability, actual power-loss,
  operational clone/rebind/rollback, deployment, Home/plugin/other-store acceptance
  or full T01/SPEC completion is claimed. Gate/query/routing/provider/hourly phases
  remain future work. The ten tests certify these specific public behaviors,
  not those later acceptance stages.


## Pass 6: existing JSON edits, owned filtered deletion and real store fixture

Date: **October 26, 2026** (Asia/Singapore). Owner-project implementation only;
all corrections remain **unstaged and uncommitted**. No nested agents, model or
route substitution, permission exception, external publication or deployment.
This is not full T01/SPEC66/70/79 acceptance or an independent review of this repair.
All earlier history, failed actions and unresolved gates above is retained.

### Frozen input and independently supplied Pass 5 evidence

- Branch `claude/spec66-resource-catalog`; HEAD/base
  `a4acc9f752bd46571f737a10c04bf413656ab06b`.
- Input: **49 staged files, +6858/-48**, worktree equal to index, no untracked files.
  The authoritative input was `git diff --cached <base> --`, not `base...HEAD`.
- Canonical raw `git diff --cached --binary <base> --` SHA256, verified at entry
  and after implementation:
  `7667dab8f4d721abe50d6127458948c3731125c08129107ce5a2fc67b99df43b`.
- Read owner AGENTS/CLAUDE, actual README contribution guidance, the saved complete
  public SPEC66/ticket70/SPEC79 bodies/comments, owner handoff/identity plan,
  catalog contract, Pass 5 verification and relevant parser/save/clone/manual
  HTTP/projection/catalog/watcher/auth-manager code and tests.
- The supplied latest independent **Spec** report found **two static P2s**:
  real JSON round-trips lose JSON-excluded provenance; a successful owned filtered
  transition then deletion silently retains A. Neither was reproduced by that
  static review. The separate **Standards** report found **one hard fixture
  violation / zero heuristics**: successful no-op `inventoryMemoryStore`.
- Supplied coordinator execution on this exact frozen input, **not execution by
  this repair worker and not evidence against the new findings**:
  - `test ./internal/api -run '^TestResourceInventory(OwnedSaveContinuity|ManualSaveContinuity)$' -count=1 -timeout=90s`:
    exit 0; api **4.943s**.
  - `test ./internal/config ./internal/api ./internal/resourcecatalog ./internal/watcher/... -count=1 -timeout=180s`:
    exit 0; config **6.621s**, api **43.723s**, catalog **4.635s**, watcher
    **6.811s**, watcher/diff **3.324s**, watcher/synthesizer **3.609s**.
  - `vet ./internal/config ./internal/resourcecatalog ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...`:
    exit 0; empty diagnostics.
  - `test -race ./internal/config ./internal/api ./internal/resourcecatalog -run '^(TestV8|TestResourceInventory(OwnedSaveContinuity|ManualSaveContinuity|OwnedSaveProjectionLifecycle|NativeEndpointClear|OwnedSaveFilteredTransition|EffectiveCredentialContinuity|EffectiveCredentialDefaults|NativeEndpointConflict|OwnedSaveWriteFailure|ManualCredentialChanges|CompatibilityFilteredProviderSave|FilteredMemberProjection|FilteredNullInheritance|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)$|TestResourceIdentity(CASAndObservationFences|ExplicitRebind)$)' -count=1 -timeout=360s`:
    exit 0; config **11.958s**, api **270.898s**, catalog **3.141s**.
    Focused only: not middleware, full SDK, full packages or full-repository race.

### Actual execution boundary and sequential red/green receipts

Every Go argv below used this exact, unchanged approved prefix:

```powershell
& 'C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9\go.ps1' -C 'D:\WILL\AGENT\CPA\CLIProxyAPI-source'
```

The existing helper was read, reused unchanged and checked as SHA256
`b33198b4c339f6a586880ed895fb3b10b207a1d78dc3117433a1964b9d50a609`.
It uses the existing task-local GOCACHE/GOTMPDIR, GOPROXY=off, GOSUMDB=off,
GOTOOLCHAIN=local and existing process GCC; GOMODCACHE was unchanged. No new
baseline/cache/evidence root was created. Receipts were captured in command stdout.
Only synthetic temp config/catalog/auth sources and controlled in-process local
upstreams were used; no normal service startup, live 8317, providers, OAuth,
browser, personal credentials/config/data or yosef-server access.

1. **P2: existing manual JSON round-trip.**

   ```text
   test ./internal/api -run '^TestResourceInventoryManualJSONRoundTrip$' -count=1 -timeout=90s
   ```

   Genuine initial assertion red: exit **1**, api **4.177s**. Both Gemini and
   compatibility existing GET arrays were decoded through real JSON, edited and
   PUT back without private fields; both returned **500**, requiring manual-save
   provenance. Native manual-off uses the existing wildcard exclusion operation;
   compatibility uses its existing `disabled` flag.

   Same-selector green after repair and the fixture refinements below: exit **0**,
   api **4.590s**. Final coverage also includes native Meta with genuinely filtered
   keyless siblings, both accepted PUT shapes (array and `items`), independent
   native exclusion/provider disablement, subsequent PATCH including JSON child
   replacement, reorder, key replacement, new native/compatibility entries and
   compatibility providers, represented deletion/delete-all, and preserved
   external filtered groups. GET/YAML do not expose transient provenance.

   The shared config persistence boundary restores only unique source addresses
   from the same writer's document-fenced, family/parent-scoped receipt. The
   compatibility GET provider ID omission is handled through proven child-parent
   receipts, not name/key/endpoint/position matching. Typed saves remain strict.
   There is no issuance/identity-validity bypass or endpoint/JSON shape change.
   Address recovery does not certify ledger issuance: an already ambiguous source
   remains ambiguous under the unchanged catalog rules. Duplicate child addresses
   cannot uniquely restore private provenance through the existing JSON shape;
   no key/name/position fallback or new API was invented to claim continuity.

2. **P2: delete after an owned transition to runtime-filtered.**

   ```text
   test ./internal/api -run '^TestResourceInventoryOwnedFilteredDeletion$' -count=1 -timeout=90s
   ```

   Genuine initial assertion red: exit **1**, api **4.336s**. After real management
   HTTP, the same Codex/compatibility writer cleared A's endpoint and saved, then
   removed only A while B remained. Save succeeded but disk retained **4 groups
   instead of 3**. Reparsed disk had forgotten that A was formerly represented.

   Same-selector green after repair and the inherited-header fixture refinement
   below: exit **0**, api **4.487s**. The final selector additionally covers both
   families using either the same writer or a fresh clone after the filtered
   transition. It tests a real temp read-only-file write failure before deletion,
   unchanged bytes and successful subsequent deletion, only B reaching the real
   watcher/auth manager, stale independent older-clone rejection, subsequent HTTP
   delete-all, fresh empty clone saves, and preservation of never-represented
   external filtered rows/empty groups.

   Immutable represented-row receipts now survive successful filtered transitions;
   regrouping and its unchanged fast path compare that owned basis rather than
   treating a fresh filtered disk parse as deletion authority. Emitted bindings,
   document fence and receipt are prepared before I/O but installed only after
   successful persistence. Older independent clones retain their old basis.

3. **Standards: real unsupported adapter, no functional-red claim.**

   ```text
   test ./internal/api -run '^TestResourceInventoryUnsupportedStorage$' -count=1 -timeout=90s
   ```

   Exit **0**, api **4.487s**, with all original HTTP503 and unsupported-storage
   inventory/durable/gate/refusal assertions retained. Removed the no-op fixture;
   used production `store.NewGitTokenStore("", "", "", "")` and only `SetBaseDir`
   pointing to synthetic temp local state. Read constructor/path setter,
   `coreauth.NewManager`, handler initialization and storage-mode checks first:
   these setup paths do not initialize or access a repository/network. No
   Load/persistence operation on this unsupported manager/store, EnsureRepository,
   clone, pull,
   push or remote call. The original harness still loads its real temp file store.
   This is fixture-quality evidence, not a new production behavior red/green.

### Supporting counterevidence and retained intermediate failures

- `TestResourceInventoryManualJSONRefusal` drives real JSON/HTTP refusals for
  unknown IDs, copied addresses, mixed compatibility parents and stale source
  documents, checking exact unchanged disk bytes. It does not weaken the prior
  typed missing/unknown/group-child/cross-binding refusal tests.
- `TestResourceInventoryManualJSONRuntime` changes a known credential through
  JSON/HTTP, checks actual catalog incarnation/identity/credential rotation,
  reparses disk, uses real watcher synthesis/auth manager/executors and observes
  exactly the replaced synthetic credential at the controlled local upstream for
  native Gemini and compatibility. Later native wildcard exclusion and provider
  disablement reach real synthesis without upstream requests. Query/gate remain
  false. This does not certify a future routing gate or provider-query phase.
- An intermediate first-finding run exited **1**, api **4.331s**: a new test
  incorrectly assumed a keyless row was runtime-filtered in every family and
  compared the last array position; compatibility also hit a nil marker assertion.
  Fix preparation also had to carry the recovered provider marker into the
  writer's post-save binding receipt, because compatibility GET omits that marker.
  The following run exited **1**, api **4.248s**, on the two overstrict keyless
  assertions. Tests were corrected to preserve/check the actual keyless source,
  retain the deliberately unremoved child, tolerate genuinely unmarked new rows,
  and cover Meta's real filtered-child behavior. Original 500 assertions stayed.
- A supporting-test command exited **1** at compile time, before tests, because
  a new local string variable was used for the typed translator Format. Replaced
  it with the existing typed executor Options value; no production change or
  functional-red claim. Exact argv:
  `test ./internal/api -run '^TestResourceInventory(ManualJSONRoundTrip|ManualJSONRefusal|ManualJSONRuntime|ManualProjectionRefusal|OwnedSaveWriteFailure)$' -count=1 -timeout=90s`.
  Corrected run: exit **0**, api **9.982s**.
- An intermediate second-finding run exited **1**, api **4.294s**: compatibility
  passed, Codex had the correct **3 groups**, but the new fixture expected identical
  raw YAML after the existing inherited-header materialization. The synthetic
  external filtered key was made explicitly self-contained (child headers), so
  unchanged-source comparison tests the intended deletion/preservation contract
  rather than forbidding existing equivalent inheritance materialization. No
  prior test or inheritance behavior was changed; the A-removal assertion remained.
- Read-only inspection also encountered Windows rg path-glob/missing-file errors.
  These were not permission denials, executed tests or assertion-red evidence.
  There was no newly denied operation or denied-action workaround.

Final focused sequential regressions, exit **0**, api **11.940s**:

```text
test ./internal/api -run '^TestResourceInventory(ManualJSONRoundTrip|ManualJSONRefusal|ManualJSONRuntime|OwnedFilteredDeletion|UnsupportedStorage|ManualProjectionRefusal|OwnedSaveWriteFailure)$' -count=1 -timeout=90s
```

### Final package gates

Complete related packages, exit **0**:

```text
test ./internal/config ./internal/api ./internal/resourcecatalog ./internal/watcher/... -count=1 -timeout=180s
```

Config **4.064s**, api **49.491s**, catalog **3.365s**, watcher **5.092s**,
watcher/diff **2.266s**, watcher/synthesizer **2.752s**. This includes existing
publication/recovery/watcher/genesis/manual YAML corrections, filtered
cross-binding/deletion/null/default preservation, Pass 5 projection refresh,
endpoint clear/default/conflict/effective credential seals, clone/opt-out behavior
and the new sequential HTTP/deletion/real-adapter tests. No existing behavioral
test was weakened; only the specified old fake fixture was replaced.

Vet, exit **0**, empty diagnostics:

```text
vet ./internal/config ./internal/resourcecatalog ./internal/watcher/... ./internal/api/... ./sdk/auth ./sdk/cliproxy/...
```

Focused changed-package race, exit **0**:

```text
test -race ./internal/config ./internal/api ./internal/resourcecatalog -run '^(TestV8|TestResourceInventory(ManualJSONRoundTrip|ManualJSONRefusal|ManualJSONRuntime|OwnedFilteredDeletion|UnsupportedStorage|OwnedSaveContinuity|ManualSaveContinuity|OwnedSaveProjectionLifecycle|NativeEndpointClear|OwnedSaveFilteredTransition|EffectiveCredentialContinuity|EffectiveCredentialDefaults|NativeEndpointConflict|OwnedSaveWriteFailure|ManualCredentialChanges|CompatibilityFilteredProviderSave|FilteredMemberProjection|FilteredNullInheritance|ManualProjectionRefusal|CompatibilityProjectionEdits|ManualProjectionDeleteAll|InheritedManualClears)$|TestResourceIdentity(CASAndObservationFences|ExplicitRebind)$)' -count=1 -timeout=360s
```

Config **13.467s**, api **301.708s**, catalog **2.477s**. This is only the exact
focused selector above, not full changed-package, middleware, full SDK or
full-repository race acceptance. Neither final package/vet/race gate failed or
was rerun without a code change. The targeted assertion reds and new-fixture
failures are separately retained above, not overwritten by final passes.

### Final freeze, scope and remaining gates

All nine changed Go paths were formatted with `gofmt -w` before their green
checks. Final read-only `gofmt -l` returned no paths; worktree and staged
`git diff --check` passed. All eleven repair paths are UTF-8 without BOM,
LF-only. Final branch/HEAD, staged 49-file +6858/-48 diff and raw binary-diff
SHA256 remain exactly the frozen input above; no untracked files or observed
concurrent drift. No staging, index/ref/branch mutation, commit/amend, fetch/push,
publication or deployment occurred. Only these repair paths differ from index.

The repair changes only these paths, all unstaged relative to the immutable index:

- `internal/config/config.go`
- `internal/config/config_load.go`
- `internal/config/parse.go`
- `internal/config/config_yaml.go`
- `internal/config/config_v8.go`
- `internal/config/resource_binding.go`
- `internal/api/handlers/management/handler.go` (shared persistence only)
- `internal/api/resource_inventory_followup_test.go` (appended sequential tests)
- `internal/api/resource_inventory_test.go` (specified real-adapter fixture only)
- `docs/resource-catalog-v1.md`
- `docs/resource-catalog-v1-verification.md`

No deprecated v0 endpoint handler/routing feature was edited. No new global
serializer/registry, dependency, identity issuance exception or business test-only
hook/collaborator was introduced. The save remains caller-serialized with the
existing management mutex; this is not a transaction with arbitrary external
source writers. No concurrency test was added, recreated or executed.

Retained open boundaries, not retried or relabeled as successes:

- Clean `cmd/server` build/run: **blocked**, unchanged shared stat-cache denial
  documented above. No retry or alternate flags/GOMODCACHE/security/cache/context.
- Protected old scratch cleanup: not touched, reused or deleted; all other
  evidence retained. Baseline task remains **ZERO tests**, with its TEMP denial,
  archive exit 128/missing object and lazy-fetch-off condition retained above.
  No baseline preparation/materialization, relocated write/cache or fetch.
- Rejected combined extra concurrent-manual addition/selector: rejected before
  execution, no test created, zero tests run. That action was not retried or
  recreated under another name. The abandoned Node spawnSync Git EPERM checker
  was not retried; only the existing owner-shell read-only Git was used.
- Historical three SDK failures and unchanged middleware race remain unresolved;
  no full SDK rerun, unrelated repair or baseline-only attribution.
- No full-repository suite/race, independent review of this unstaged repair,
  Linux runtime/durability, physical power-loss acceptance, operational
  clone/rebind/rollback, Home/plugin/other-store or deployment acceptance.
  Gate/query/routing/balance/provider/hourly phases remain **future work**.


## Pass 7c follow-up receipt (2026-10-26, owner-local Windows run; incomplete SDK slice)

### Scope, explicit authorization and telemetry isolation

The user explicitly selected task-only telemetry OFF for this task, not a generic
continuation grant. No global telemetry directory/settings/ACLs were accessed or
changed. No `go telemetry` or `go env -w` command was run, and no public
non-settable `GOTELEMETRY`/`GOTELEMETRYDIR` override was fabricated. No model,
routing policy, credentials, service, provider, network or baseline operation was
performed. The implementation owner ran checks directly without child agents or
Codex invocations.

Existing evidence root (all paths below are relative to this root unless stated):

```text
C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9
```

A new directory was created exclusively, with no overwrite:
`telemetry-off-pass7c-20261026`. Its only file, `mode`, contains exactly
`off 2026-10-26` (14 bytes, no newline), SHA256
`3a28ae910769fee6d2be951fb9d8c5d980e67a30e27c9e74fe9cceefd2c0ed4f`.
Before and after EACH of the three new verification processes, the mode bytes
and hash were checked and the directory contained only this unchanged file.
No local/upload/debug/token/counter/report files were observed there.
`GO_TELEMETRY_CHILD=1` was checked and was not inherited. The sole added
process-only verification environment override was
`TEST_TELEMETRY_DIR=<ROOT>\telemetry-off-pass7c-20261026`; descendants inherited it.
No environment change was persisted.

The actual installed Go source was inspected independently, read-only:

- `C:\Program Files\Go\src\cmd\internal\telemetry\telemetry.go:32-35,44-47`
  passes `TEST_TELEMETRY_DIR` into Start/MaybeChild.
- `cmd/internal/telemetry/counter/counter.go:22-25` passes it to OpenDir;
  vendored `golang.org/x/telemetry/counter/counter.go:109-113` selects NewDir
  before internal counter.Open.
- Vendored `internal/counter/file.go:377-389` returns on OFF before opening
  counter files.
- Vendored `start.go:44-48,144-173` documents isolated testing and selects the
  requested directory, then returns on OFF before counter.Open, local stat,
  upload-token acquisition or sidecar start.
- Vendored `internal/telemetry/dir.go:36-47,115-149` constructs paths without
  creating directories and reads/trim-parses the dated OFF mode.

Exact focused excerpts and full source-file hashes are retained in
`pass7c-source-proof.json`. This is source-based off-short-circuit evidence plus
controlled mode/directory observations, NOT a network trace.

The existing `go.ps1` was not edited (SHA256
`b33198b4c339f6a586880ed895fb3b10b207a1d78dc3117433a1964b9d50a609`).
It retained its original GOCACHE=`<ROOT>\cache`, GOTMPDIR=`<ROOT>\build`,
GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local and existing GCC/PATH setup.
GOMODCACHE and global settings were unchanged; no alternate cache was created.
The task-only mode and receipt helper are evidence files, not project changes.

### Commands and retained execution distinctions

All Go test argv below have the exact prefix:

```text
powershell.exe -NoProfile -File <ROOT>\go.ps1 -C D:/WILL/AGENT/CPA/CLIProxyAPI-source
```

Middleware tail (identical in the historical RED, historical exit-0-with-denial,
and new clean check):

```text
test -race ./internal/api/middleware -run ^TestRequestLoggingMiddleware_StreamingPreservesFullUUIDForLogger$ -count=10 -timeout=180s
```

SDK original normal-scheduler tail:

```text
test ./sdk/cliproxy -run '^(TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch|TestEndToEndAuthFileReplacement_RestoresModelsInV1ModelsWithoutRestart|TestEndToEndAuthFilePatch_RestoresModelsInV1ModelsWithoutRestart)$' -count=10 -v -timeout=180s
```

The authorized bounded original SDK counterpart has the SAME argv with a final
`-cpu=1`. Exact argument arrays (no shell-quoting ambiguity), sole scoped extra
environment override, exit/timing, unfiltered complete stdout/stderr and SHA256
are retained as `<prefix>.argv.json`, `.result.json`, `.stdout.log`, `.stderr.log`.
Historical Pass 7b files were read and retained, not overwritten or reproduced;
`pass7c-retained-pass7b-hashes.json` captures their file hashes.

| Prefix | Exit | Wall seconds | Meaning |
| --- | ---: | ---: | --- |
| `pass7b-middleware-red` | 1 | 30.3828871 | Historical actual RED: processStreamingChunks read / Finalize channel-clear race |
| `pass7b-middleware-green` | 0 | 11.4962568 | Historical exit 0 but stderr upload-token Access is denied; NOT clean acceptance |
| `pass7c-middleware-clean` | 0 | 27.2961963 | New clean race check under explicitly changed task-only telemetry environment |
| `pass7c-sdk-original-normal` | 0 | 27.3143928 | Unmodified SDK original normal scheduler: all three tests passed 10 times each |
| `pass7c-sdk-original-cpu1` | 0 | 15.1512499 | Unmodified SDK original cpu=1: all three tests passed 10 times each |

Exact retained timings and output signatures:

- `pass7b-middleware-red`: 2026-10-26T12:16:20.4982836+08:00 -> 2026-10-26T12:16:50.8877586+08:00.
  stdout SHA256 `021d23441d02e755057550a680f00e70af3eeca5ddf871e98e1d56355d3f1f94`;
  stderr SHA256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

- `pass7b-middleware-green`: 2026-10-26T12:17:17.7238096+08:00 -> 2026-10-26T12:17:29.2250927+08:00.
  stdout SHA256 `70bc7c252b26c78b1c03ddc3af688e6ad0c4bac343e7d976d3216c2975879bf0`;
  stderr SHA256 `9b9097b7d070c430564297e3b831f60f5a7a40cfb93416cfe6571a6c35e6014a`.

- `pass7c-middleware-clean`: 2026-10-26T12:41:46.083621+08:00 -> 2026-10-26T12:42:13.379477+08:00.
  stdout SHA256 `7b4f67b7e7f3c22a856e4931cd8eeeee4854a65556617ac5c315f28023160654`;
  stderr SHA256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

- `pass7c-sdk-original-normal`: 2026-10-26T12:42:22.244671+08:00 -> 2026-10-26T12:42:49.558414+08:00.
  stdout SHA256 `719abe0a6062b32763a74c550e88f23ae3670b20364865419c9697c1bef067c9`;
  stderr SHA256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

- `pass7c-sdk-original-cpu1`: 2026-10-26T12:43:19.116682+08:00 -> 2026-10-26T12:43:34.267597+08:00.
  stdout SHA256 `a672f9278479a9c5aeadf6b27f44d6128890c31da6308ee5995b0753b1ef07b5`;
  stderr SHA256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

The new middleware check had empty stderr and package time 4.001s. Its complete
stdout was inspected as well. The pre-existing minimal production repair was
preserved byte-for-byte: close chunkChannel, wait for streamDone, only then clear
chunkChannel. No rollback was used to manufacture a new RED. These are three
distinct executions, NOT a paired original-base comparison.

### SDK evidence gate and deliberately incomplete work

Before considering edits, the actual test, handleAuthUpdates, runtimeAuthSyncHook,
model-registration task execution and per-auth completion contracts were read.
The current test blocks the second no-argument modelRegistrationTaskHook call;
that call carries no auth identity and concurrent workers need not execute in
input order. On failure its cleanup releases the blocked hook without joining
the batch and duplicate-update goroutines before clearing the hook and
unregistering the global registry clients. These are inspected potential test
synchronization/cleanup mechanisms, NOT proof of baseline-only causality or of
the historical replacement/patch model failures.

Both authorized bounded ORIGINAL SDK commands passed: 30 named test passes per
run, no failed assertions, empty stderr, package times 7.979s (normal) and
10.751s (cpu=1). Thus NO functional SDK RED was obtained. No synchronization,
cleanup, fixture, assertion, production hook, low-level seam or SDK production
code was changed; no deterministic new regression was introduced without an
established failing case. A-not-waiting-for-B, terminal-error model absence and
actual replacement/patch recovery assertions remain unchanged. This is a
non-reproduction receipt, NOT an SDK RED-to-GREEN repair claim. The SDK slice
remains incomplete under the specified gate.

### Checks not run, integrity and remaining gates

Because both repaired slices did not succeed, Phase 3 was not entered:
no gofmt, focused vet, full two-package non-race suite or named SDK race run was
performed. No Go file was newly modified in Pass 7c. In particular, an SDK
original pass is not relabeled a post-repair GREEN. No full-repository tests,
cmd/server build/EXE, new frozen candidate or acceptance action was performed.

Read-only start integrity matched the dispatched branch
`claude/spec66-resource-catalog`, HEAD/base
`a4acc9f752bd46571f737a10c04bf413656ab06b`, 51 staged files +7751/-51,
cached raw binary diff SHA256
`eebc1ae61b90da02c1293f09a2bac129991339b5d9644f7efbc493cf9118a777`,
raw index SHA256
`01e11b190172f6af2922903329519a73c99e2ae0fb6cb8f9ad98da620ff3ef46`,
and raw for-each-ref stdout SHA256
`c8a30bba3783770da61923945250c55d32700a718afd8c8509672c17949e6410`.
Read-only after-integrity is recorded in `pass7c-after-integrity.json`, using
GIT_OPTIONAL_LOCKS=0 and GIT_NO_LAZY_FETCH=1. Expected full unstaged path list:
`internal/api/middleware/response_writer.go` (preserved prior repair) and
`docs/resource-catalog-v1-verification.md` (this append only). No untracked
project files are expected. Worktree and cached whitespace-check outputs/exits
are retained with the after-integrity receipt. Original documentation bytes
are prefix-verified, preserving all historical failures/stopped attempts.

No staging/index/ref/branch/worktree creation/commit/fetch/push, protected
scratch access/cleanup, rejected concurrent-manual test, baseline retry,
external publication, live service or global telemetry access occurred.
The original private Pass 6 backup was not touched. No new permission denial
occurred in these three task-isolated checks; the historical denial remains
retained and is not erased by the new explicit grant.

Remaining owner/parent work: obtain a genuine SDK regression at an approved seam
before any bounded synchronization/cleanup repair; then exact failed-command and
normal-scheduler GREEN, named SDK race, formatting, focused vet and complete
non-race tests of the two repaired packages. Broader package/race tests, clean
cmd/server build, a new frozen candidate and independent standards/spec reviews
remain pending. Baseline preparation remains stopped before process creation
(objects 0, matrix 0/4), with no baseline retry here. No local check establishes
upstream-baseline causality, Linux/durability/runtime/UI or full SPEC66/79
acceptance. Historical blocked gates remain blocked, not silently promoted.


## Pass 7d owner validation receipt (2026-10-26; stopped on original SDK functional RED)

### Scope and preserved starting state

This was direct owner validation of the actual middleware-only repaired worktree,
not a routing task, an SDK repair or a staged freeze. No nested Codex process,
agent, route validator, settings/security/permission change, provider/service
operation or baseline retry was used. The existing Pass 6 and Pass 7b/7c history
is preserved. In particular, the clean Pass 7c middleware race execution and its
SDK normal/cpu=1 non-reproductions are not rewritten or promoted to broader
acceptance.

Read-only starting checks used GIT_OPTIONAL_LOCKS=0 and GIT_NO_LAZY_FETCH=1:
branch `claude/spec66-resource-catalog`, HEAD/base
`a4acc9f752bd46571f737a10c04bf413656ab06b`; original cached change set remained
51 files, +7751/-51. The raw identities matched the dispatched values:

- Cached binary diff against the pinned base: SHA256
  `eebc1ae61b90da02c1293f09a2bac129991339b5d9644f7efbc493cf9118a777`.
- Index file: SHA256
  `01e11b190172f6af2922903329519a73c99e2ae0fb6cb8f9ad98da620ff3ef46`.
- Raw `git for-each-ref --format=%(refname) %(objectname)` stdout: SHA256
  `c8a30bba3783770da61923945250c55d32700a718afd8c8509672c17949e6410`.

Exactly two unstaged paths existed: `internal/api/middleware/response_writer.go`
and this verification document. No untracked project artifacts existed.
`pass7d-start-integrity.json` retains these checks, dirty-path file hashes and
378 SDK file hashes. The entire pre-append document was saved externally as
`pass7d-doc-prefix.bin` (87296 bytes, SHA256
`c8afbf00758d43c0241d462cf99a92e06c92d19adc22d70bb343592bf52dfc42`), and the
existing Pass 7c append was confirmed to be exactly 185 lines beyond the index.

All new evidence is exclusive `pass7d-*` material under the EXISTING root:

```text
C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9
```

No old receipt, private Pass 6 backup, cache location or original runner was
replaced. No staging/index/refs/branch/commit/push/deploy action occurred.

### Existing task-only telemetry OFF environment

The existing `telemetry-off-pass7c-20261026` directory was reused without any
creation or modification. Before and after the sole Go invocation it contained
ONLY `mode`, exactly 14 bytes `off 2026-10-26`, with SHA256
`3a28ae910769fee6d2be951fb9d8c5d980e67a30e27c9e74fe9cceefd2c0ed4f`.
GO_TELEMETRY_CHILD=1 was not inherited. The process-only telemetry override was
`TEST_TELEMETRY_DIR=<ROOT>\telemetry-off-pass7c-20261026`; the runner and its
descendants inherited it. Git read-only flags were also process-scoped.

The unchanged original `go.ps1` SHA256 was
`b33198b4c339f6a586880ed895fb3b10b207a1d78dc3117433a1964b9d50a609` before and
after execution. It retained GOCACHE=`<ROOT>\cache`, GOTMPDIR=`<ROOT>\build`,
GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local and its existing local GCC/PATH.
These existing caches were reused normally, not relocated or replaced;
GOMODCACHE remained inherited and unchanged. No alternate cache was created.
No global telemetry/config/token/counter directory was inspected or changed;
no Go telemetry command, GOTELEMETRY override or persistent environment change
was used.

The six shipped-source hashes in the retained `pass7c-source-proof.json` were
read back and verified unchanged; `pass7d-source-proof-readback.json` records
this check. That retained source proof describes the TEST_TELEMETRY_DIR shim,
OpenDir selecting the supplied directory before counter.Open, and OFF returns
before counter writes, local stat, upload-token acquisition and sidecar start.
The unchanged mode/sole-file observations and source proof are NOT a network
trace. No new permission/filesystem/environment denial was observed; the earlier
telemetry denial history remains preserved.

### Sole executed Go validation and actual failure

Exact command (the quoted selector is one argv element; complete argv and
process-only overrides are retained in the receipt):

```text
powershell.exe -NoProfile -File <ROOT>\go.ps1 -C D:/WILL/AGENT/CPA/CLIProxyAPI-source test -race ./sdk/cliproxy -run '^(TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch|TestEndToEndAuthFileReplacement_RestoresModelsInV1ModelsWithoutRestart|TestEndToEndAuthFilePatch_RestoresModelsInV1ModelsWithoutRestart)$' -count=10 -v -timeout=180s
```

Receipt prefix: `pass7d-sdk-original-race`. The invocation started
`2026-10-26T12:56:39.765635+08:00` and finished
`2026-10-26T12:58:12.032106+08:00`; actual runner/Go exit was **1**, wall time
**92.268872300 seconds**. Package output reported **47.632s**. All 30 named
instances completed within that one invocation: **27 PASS, 3 FAIL**, with each
of the three named tests passing nine times and failing once. No retry,
increased count or CPU stress matrix was run.

Complete, unfiltered process stdout (17964 bytes) and stderr (0 bytes), along
with exact argv, timing, exit, runner hash, before/after mode observations and
source hashes, are retained in `.stdout.log`, `.stderr.log`, `.argv.json` and
`.result.json`. Their SHA256 values are:

- stdout: `1ef12c20a6fb37cfe84f1ab0a7cd57b2546128c35c7726f2d7277b89f8ab0100`.
- stderr: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

Actual failure diagnostics, all in stdout:

```text
service_auth_sync_test.go:763: auth A hook wait blocked on unrelated auth B registration
service_auth_sync_test.go:1397: expected "gpt-6-astra" to be absent after terminal auth failure
service_auth_sync_test.go:1557: expected "gpt-6-astra" to be absent after terminal auth failure
```

`pass7d-failure-summary.json` retains the exact stdout line references, test
occurrences and outcome counts. The assertions failed sequentially in the
original chain. There was **no `WARNING: DATA RACE`** in the output: this is a
functional RED observed under `-race`, not a reported data race or permission
denial. Sequential adjacency does not prove contamination, hook identity,
root cause or baseline-only causality. No paired baseline exists here; the
historical failures remain unattributed. The no-identity func() registration
hook's invocation order is not treated as auth B identity.

SDK production code, tests, fixtures, assertions, hooks and API seams were NOT
modified. Pass 7c had no reproduced SDK RED; Pass 7d now has the actual original
functional RED above. Accordingly, it would be false to describe this task as
having no reproduced RED. This is evidence for a subsequent bounded owner
repair, **not an SDK repair or RED-to-GREEN claim**. No manufactured regression,
fixture reversal, fake scheduler or weakened assertion was introduced.

### Ordered stop, intentionally skipped checks and remaining gates

After the failed first command, no further Go/gofmt/vet/build work ran:

| Ordered step | Actual status |
| --- | --- |
| 1. Original named SDK chain, race/count=10 | FAILED as recorded above; sole Go invocation |
| 2. gofmt only response_writer.go | SKIPPED after functional failure; no formatting write |
| 3. Complete related-package non-race check | SKIPPED; no package selector substitution |
| 4. Related vet coverage | SKIPPED |
| 5. Fresh required cmd/server compile | SKIPPED; target did not preexist and was not created |
| 6. Append factual English receipt | This append only; entire previous prefix preserved |
| 7. Read-only final integrity/candidate identity | Retained externally in pass7d-final-integrity.json |

No fresh `cli-proxy-api-pass7d-20261026-check.exe` exists and no executable was
run. The prior Pass 6 executable is untouched and is not build evidence for
this worktree. The middleware close -> worker join -> channel nil repair and
its English comment were preserved byte-for-byte; source SHA256 before and
after the failed command was
`3e7df65c701fd0b1af6ee8a2c3d7a7737486830c9107d12d9845ca830c140b55`.
There is no post-format hash because formatting did not run.

The external final readback records original cached/index/refs/runner checks,
full unstaged/untracked lists, SDK byte equality, document prefix preservation,
whitespace command outputs/exits, source/final document hashes, and the raw
`git diff --binary a4acc9f752bd46571f737a10c04bf413656ab06b --` working-tree
candidate identity/stat/count. The candidate includes staged and unstaged
changes and is NOT a staged freeze or commit. No candidate hash is embedded in
this document, avoiding self-reference. Base...HEAD remains empty.

The SDK RED requires a separately bounded owner repair before resuming these
ordered verification gates. Independent Standards and Spec static reviews,
formatting, related tests/vet, a fresh required build, staging/commit and all
other unfulfilled acceptance gates remain pending. The baseline attempt is
still unresolved at objects 0, matrix 0/4; no retry or baseline attribution was
made. No full-repository, Linux, durability, clone/rollback/power-loss, runtime,
UI or full SPEC66/79 acceptance is claimed. Historical limitations are not
weakened by this failure receipt. The completed verification subprocess was
joined; no validation task remains running at handoff.


## Pass7e - bounded owner test coordination repair (2026-10-26)

### Scope and preserved original RED

This pass directly repaired only the coordination and cleanup inside
`TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch`, plus
its required `sync` import. The actual `handleAuthUpdates` / `handleAuthUpdate`
SDK seam and the existing no-argument `modelRegistrationTaskHook` remain in
use. No SDK production source, hook signature, collaborator, auth/provider
fixture, classification, or business behavior was changed. This was one
coherent test-only edit; no additional repair iteration was needed.

Starting integrity matched branch `claude/spec66-resource-catalog`, HEAD/base
`a4acc9f752bd46571f737a10c04bf413656ab06b`, the pinned cached diff, index,
refs, runner, full working-tree diff and both existing unstaged files.
`pass7e-start-integrity.json` records the actual hashes and the original RED
readback. Existing private backups, receipts, index, refs, and history were
not rewritten. The baseline remains blocked at objects 0 and matrix 0/4.

The preserved `pass7d-sdk-original-race` command was not rerun or manufactured:
actual exit **1**, wall **92.2688723s**, package **47.632s**,
**27 PASS / 3 FAIL** among 30 named cases, with each of the three failures on
its seventh occurrence. The diagnostics were:

- original line 763: `auth A hook wait blocked on unrelated auth B registration`;
- original lines 1397 and 1557: `expected "gpt-6-astra" to be absent after terminal auth failure`.

That was functional RED under race instrumentation, not a detected data race:
no `WARNING: DATA RACE`, no permission/environment denial. Its stdout SHA256
is `1ef12c20a6fb37cfe84f1ab0a7cd57b2546128c35c7726f2d7277b89f8ab0100`;
empty stderr SHA256 is `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
The original SDK file hash was
`f41aed1cbbb0ac194951e5b83e052fa36e2ecc441f23bb4eb99a583355700333`.
Historical original normal/cpu1 passes are non-reproductions, not repair
proof. Neither the original failure sequence nor subsequent GREEN establishes
baseline attribution or proves downstream registry pollution causality.

### Identity-safe scheduling, assertions, and cleanup

The existing hook runs before `task.run` in parallel registration workers;
its invocation ordinal does not identify an auth. Both original tasks now
publish separate private gates and remain blocked before either registration
can execute. Same-revision duplicates are launched for both actual update
IDs, and each completion carries that update's real ID. Releasing exactly
one gate lets the completed duplicate identify the ready auth without
assigning A/B from worker order. The test validates its ID, requires the
other duplicate and original batch to remain incomplete while the second
gate stays closed, then releases that gate and requires the distinct other
auth ID and the batch to finish. It also rejects duplicate completion before
either original gate is released. The semantic per-auth independence
assertion is preserved, with explicit blocked-other/batch and identity checks.
No private per-auth wait-map inspection, new business seam, fake collaborator,
wall-clock sleep, fixture reclassification, or weakened absence/recovery
assertion was introduced. Existing two-second waits are deadlock safety
bounds rather than scheduling sleeps.

Completion-aware cleanup is installed before launching callers. Its shared
release is idempotent (`sync.Once`); both arrival publication and gate waits
select on this release, including unexpected hook invocations. Buffered
completion channels permit every duplicate to exit even after a fatal
assertion. The cleanup releases all hooks, joins the original batch and both
duplicates through a `WaitGroup`, and only then clears the hook. Because this
cleanup was registered after the existing registry cleanup, LIFO execution
joins callers before unregistering either client. The batch itself waits for
its registration workers. Thus cleanup ownership no longer permits these
callers to register after client cleanup. This is code-path evidence;
no extra forced-fatal experiment was run, and arbitrary production deadlocks
are still bounded by the Go test timeout rather than claimed impossible.

`pass7e-scope-proof.json` verifies that all file bytes outside this function
and the one added import are identical. In particular the complete replacement
and patch tests, including terminal-error absence and actual `/v1/models`
recovery assertions, are byte-identical:

- Replacement: 6353 bytes,
  SHA256 `9e57b7809509dc90a83fe3bf9da62f9f16b1dbf6b94e5c2d5e52e8ba05b5b897`.
- Patch: 5124 bytes,
  SHA256 `c25f7ea12abe689d15aec27bbd2c83b6685ee3f9e4bc67d1153b4fa586dcf2a6`.

### Actual commands and results on formatted source

All receipts below are in the existing external directory
`C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9` (abbreviated `R`).
Every prefix has new `.argv.json`, `.result.json`, complete unfiltered
`.stdout.log`, and `.stderr.log` files. Results record actual process exit,
start/end/wall timing, stdout/stderr sizes and hashes, source hashes, runner
hash, scoped environment, and before/after OFF directory observations.
The recorded child exits and package/test results, not a wrapper exit alone,
are the basis of the following results. All five commands had empty stderr.

```powershell
$R = 'C:\Users\will\AppData\Local\Temp\cpa-spec66-followup-db557e1f236f494aa64dfa28b69fd9c9'
$S = '^(TestHandleAuthUpdates_SameRevisionWaitDoesNotWaitForOtherAuthInBatch|TestEndToEndAuthFileReplacement_RestoresModelsInV1ModelsWithoutRestart|TestEndToEndAuthFilePatch_RestoresModelsInV1ModelsWithoutRestart)$'
# Each command inherits process-only TEST_TELEMETRY_DIR=$R\telemetry-off-pass7c-20261026
# and GIT_OPTIONAL_LOCKS=0, GIT_NO_LAZY_FETCH=1.
& 'C:\Program Files\Go\bin\gofmt.exe' -w sdk/cliproxy/service_auth_sync_test.go internal/api/middleware/response_writer.go
powershell.exe -NoProfile -File "$R\go.ps1" -C D:/WILL/AGENT/CPA/CLIProxyAPI-source test -race ./sdk/cliproxy -run $S -count=10 -v -timeout=180s
powershell.exe -NoProfile -File "$R\go.ps1" -C D:/WILL/AGENT/CPA/CLIProxyAPI-source test ./sdk/cliproxy -run $S -count=10 -v -timeout=180s
powershell.exe -NoProfile -File "$R\go.ps1" -C D:/WILL/AGENT/CPA/CLIProxyAPI-source vet ./internal/api/middleware ./sdk/cliproxy
powershell.exe -NoProfile -File "$R\go.ps1" -C D:/WILL/AGENT/CPA/CLIProxyAPI-source test ./internal/api/middleware ./sdk/cliproxy -count=1 -v -timeout=240s
```

| Receipt prefix | Actual exit | Wall seconds | Test/tool outcome |
| --- | ---: | ---: | --- |
| `pass7e-format` | 0 | 0.3419764 | Only the two approved Go files; neither file's bytes changed |
| `pass7e-sdk-race` | 0 | 60.9684287 | 30 PASS / 0 FAIL; each selected test 10/10; package 42.681s |
| `pass7e-sdk-normal` | 0 | 15.9992663 | 30 PASS / 0 FAIL; each selected test 10/10; package 6.581s |
| `pass7e-vet` | 0 | 4.6008799 | Focused vet of the two approved packages, no diagnostics |
| `pass7e-full-two-packages` | 0 | 21.5082735 | Both full non-race packages passed; 214 top-level PASS, 0 top-level SKIP, 0 FAIL |

Full package result lines:

```text
ok  	github.com/router-for-me/CLIProxyAPI/v8/internal/api/middleware	4.358s
ok  	github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy	12.108s
```

No race detector warning or new permission/filesystem/environment denial was
observed in these checks. The corrected race/normal command chains retained
all three selected tests. The original middleware lifecycle repair was kept:
close the chunk channel, wait for its reader's `streamDone`, then clear the
channel field. Gofmt did not change its bytes; this pass added no middleware
logic. Full non-race package success is not a new full-package race run.

Formatted SDK test SHA256:
`f105f8fe8732bd7d8efd6fc429779c23555cf05666a047272fdc3d77a5f3f499`.
Middleware SHA256 before and after formatting/testing:
`3e7df65c701fd0b1af6ee8a2c3d7a7737486830c9107d12d9845ca830c140b55`.

### Task-only OFF mode and integrity boundaries

Before and after every Go/gofmt/vet command, the same task OFF directory
contained only `mode`, exactly 14 bytes `off 2026-10-26`, SHA256
`3a28ae910769fee6d2be951fb9d8c5d980e67a30e27c9e74fe9cceefd2c0ed4f`.
`GO_TELEMETRY_CHILD=1` was not present. Every tool process inherited only the
approved process-scoped `TEST_TELEMETRY_DIR` selection, with unchanged Git
safety environment and original runner settings. The runner remained SHA256
`b33198b4c339f6a586880ed895fb3b10b207a1d78dc3117433a1964b9d50a609`.
GOCACHE stayed `R\cache`, GOTMPDIR `R\build`, GOPROXY/GOSUMDB `off`,
GOTOOLCHAIN `local`, and the existing GCC/PATH prefix was reused.
GOMODCACHE was inherited unchanged, not relocated; cache contents may
naturally change during verification and are not claimed byte-identical.
No global telemetry/config/credential survey or settings command was run.
The retained prior source proof and these directory observations are not a
network trace and do not prove impossible global activity.

This English section is append-only after the complete original 96462-byte
document prefix, SHA256
`3221c930199a49ec2feaa691b4a6fe9096a771bd19e1fc7ce32bf1392056c918`.
The final external `pass7e-final-integrity.json` records prefix preservation,
source/document hashes, full raw working-tree binary-diff SHA256/byte count
and stat versus the pinned base, cached/index/refs/runner equality, exact
unstaged/untracked paths, and whitespace-check outputs/exits. The raw
candidate diff is external; its hash is deliberately not embedded in this
hashed document. The candidate remains a working tree, not a staged freeze
or commit. Base...HEAD remains empty. Only the SDK test, pre-existing
middleware repair, and this verification document are unstaged.

### Remaining verification and handoff limits

The bounded repair and ordered focused checks above are complete. Broader
related-package/repository validation, a fresh required `cmd/server` build,
and independent Standards and Spec reviews remain for the parent. This pass
did not run a server EXE, contact live services/providers, stage/commit/push,
change model policy or global settings, create agents, or execute nested
Codex. It made no Linux/durability/clone/rebind/rollback/power-loss/runtime/UI
or full SPEC66/79 acceptance claim. Baseline evidence is still unresolved,
not baseline-only attribution. Verification subprocesses were joined before
handoff; no verification worker is intentionally left running.

## Pass8b bounded follow-up receipt (2026-10-26)

This receipt covers the three authorized follow-up blockers on Windows. It is
not full SPEC66 acceptance, an independent review, a routing validation, or a
baseline execution. The parent owns publication and independent Standards/Spec
reviews after this worker exits.

### Preservation and execution boundary

Before source edits, the worker saved the exact raw full candidate to
`<ROOT>/pass8b-preserved-candidate.patch` and checked the supplied candidate,
HEAD, branch, cached diff, index, refs, protected middleware/SDK files, runner,
and the original 107014-byte documentation prefix. Start evidence is
`pass8b-start-integrity.json`. Git output used raw byte streams, not PowerShell
text encoding. No staging, commits, ref/object writes, fetches, deployments,
external publication, tool installation, Python, or nested agents were used.

Every Go operation, including formatting through `go fmt` (which invokes gofmt
only for the named changed Go files), used the original `<ROOT>/go.ps1` runner.
Its cache/build directories, offline module settings, local toolchain and GCC
were retained. Each Go receipt records raw stdout/stderr, argv, the actual child
exit code, UTC timing, output hashes, and before/after telemetry checks. The
supplied telemetry directory contained only the exact OFF mode file; the runner
hash was checked before and after every command. These are source/environment
receipts, not a network trace. No live credential, provider, OAuth, production
service, browser, or personal settings operation was run by this worker.

### Watcher event to real reload callback

The unchanged original selector, `TestHandleEventConfigChangeSchedulesReload`,
passed ten repetitions (23.140 seconds). The original related-full command also
passed once (66.916 seconds). The supplied historical RED remains valid evidence
but was not reproduced in these two runs; no production reload cause is claimed.

Four falsifiable hypotheses were recorded before the probe: an observation
before callback completion, a lost debounce callback, a file visibility/hash
mismatch, and a config parse rejection. The test did have a provable observation
race: its fixed 400ms sleep did not synchronize with the asynchronous callback.
Only that test was changed to wait on a channel signaled by the real callback,
then retain the exact-one count assertion. No production watcher/debounce logic,
private map, fake callback path, new test hook, or timeout value was changed.
The same ten-run loop and watcher packages passed afterward. Final repeated and
race results are in the table below. Details: `pass8b-watcher-stage.json`.

### Deprecated v0 saves versus native v8 protection

An authenticated real HTTP PUT to `/v0/management/codex-api-key` reproduced the
regression: a replacement carrying a caller ID received 500 from the new
unambiguous-provenance refusal. Changing only the wrapper was insufficient: the
lower regroup path still refused it. An explicit legacy-management save path
now retains pre-catalog v0 semantics, excluding new source-projection/CAS rules.
Ordinary config writers retain their existing source protection. The real v0
HTTP regression turned green; route registrations were not changed.

Feature HTTP tests now use the existing native
`/v8/management/config/api-keys/<family>` boundary, not invented legacy routes
under v8. Migrating the tests first exposed eight real RED refusal cases: native
v8 incorrectly returned 200 for unknown IDs, copied IDs, stale source edits and
cross-provider/parent membership. Validation now lives at the actual ConfigV8
write boundary. Issued IDs remain scoped to family, kind and parent; stale
source writers are refused before persistence. Unmarked new rows can still be
enrolled. An exact source no-op remains valid after catalog marker enrollment.
The normalized saved document is rebound to the next runtime writer receipt.

The tests use native raw groups/keys rather than legacy filtered arrays or
`items` envelopes. They retain meaningful round-trip, reorder, replacement,
addition, refusal/no-write, manual ownership, filtered/keyless source,
identity/version and real watcher/manager/executor assertions. Native raw-list
deletion explicitly retains the untouched filtered source; filtered deletion,
stale clones and failed-write ownership stay tested through the existing public
config writer. Compatibility fixture headers were moved from unsupported child
headers to the provider, retaining exact filtered-source assertions instead of
weakening native schema validation.

A further real HTTP RED covered the first GET/edit/PATCH immediately after
enrollment, without the tests manually calling SetConfig: marker migration had
left the handler's source receipt stale. Successful catalog initialization now
refreshes its ResourceCatalog receipt from the current persisted source without
replacing runtime key lists or other configuration sections. First-edit and
external-stale refusal cases both passed afterward. Evidence:
`pass8b-v0-stage.json`, `pass8b-v0-enrollment-stage.json` and their command receipts.

### Typed-invalid sources must not publish ready inventory

Six real API/catalog RED cases covered grouped and native flat layouts for
sequence-valued `api-key`, `base-url`, and `headers`. Native config parsing
rejected every fixture, but reconciliation previously returned success and
published HTTP 200 with `complete=true` and ready rows. Generic YAML shape/layout
validation did not apply native types, and credential seals ignored Decode
errors, turning malformed credential fields into empty values.

Inventory enumeration now performs native Config YAML decoding before building
its proposal, without runtime sanitization or filtering. Credential Decode
errors propagate through seal calculation before any proposal publication.
The same API tests now require unavailable inventory/capabilities, no malformed
source rewrite, and successful valid-source recovery with exact prior target
and manual-state preservation. Raw disabled, keyless and zero-weight resources
remain present; runtime readiness stays unknown and gate/query stay false.
Two additional empty-provider cases cover base-url/headers validation when no
flattened runtime row exists; they exercise the seal error path and valid
identity-preserving recovery. No extra RED is claimed for this post-fix coverage.
Evidence: `pass8b-types-stage.json`, `pass8b-types-empty-stage.json`. The initial
`pass8b-types-red` receipt was a missing test-helper compilation error;
`pass8b-types-red2` is the actual six-case functional RED.

### Final formatted-code verification

Exact argv and selectors are retained in each named JSON receipt. Commands use:

```text
powershell.exe -NoProfile -File <ROOT>/go.ps1 -C D:/WILL/AGENT/CPA/CLIProxyAPI-source <args>

test ./internal/config ./internal/api ./internal/api/middleware ./internal/resourcecatalog ./internal/watcher/... ./sdk/auth ./sdk/cliproxy -count=1 -timeout=240s
test -json ./... -count=1 -timeout=240s
vet ./...
build -o <ROOT>/cli-proxy-api-pass8b-20261026-check.exe ./cmd/server
```

The changed API selector was repeated twice; watcher selectors ten times.
Focused race was split into two bounded groups with the same 360s timeout, not
inflated. The original three SDK selectors and the original middleware selector
were each repeated ten times in both ordinary and race modes. This is not a
full-repository race run.

| Receipt stem | Actual child exit | Wall seconds |
|---|---:|---:|
| `pass8b-final-focused2` | 0 | 41.231 |
| `pass8b-final-watcher` | 0 | 17.125 |
| `pass8b-final-race-a` | 0 | 150.289 |
| `pass8b-final-race-b` | 0 | 136.564 |
| `pass8b-final-sdk-normal` | 0 | 12.443 |
| `pass8b-final-sdk-race` | 0 | 52.855 |
| `pass8b-final-middleware-normal` | 0 | 7.007 |
| `pass8b-final-middleware-race` | 0 | 10.168 |
| `pass8b-final-related` | 0 | 127.651 |
| `pass8b-final-fullrepo` | 0 | 396.719 |
| `pass8b-final-vet` | 1 | 13.725 |
| `pass8b-final-build` | 0 | 7.963 |

The related-full run passed all nine package results. The full-repository JSON
run reported 100 passing packages, 27 no-test-file package skips, and no failing
packages/build failures. Test-result events, including subtests, were 16388 pass,
0 fail and 9 skip; these are not a top-level-test count inferred from quiet
output. Exact package outcomes and skipped test names are recorded in
`pass8b-fullrepo-summary.json`. Skip reasons are in
`pass8b-fullrepo-skip-reasons.json`: one opt-in live Devin test, six absent native
signature/corpus tests, one unset TLS capture proxy, and one Windows
case-insensitive environment test. These skips and uncompiled Linux paths are
not waived coverage.

A wider intermediate race selection timed out at the API package's aggregate
360s deadline (394.975 seconds including compilation). The active stack was
synthetic bcrypt computation, not a reported data race. That failure remains in
`pass8b-final-race.json`; both split final race groups passed without changing
crypto costs, assertions, or timeout values. Intermediate v0 wrapper-only and
native-fixture/schema/receipt failures remain preserved in their receipts.

**Vet remains failed (exit 1):** 13 unsafe.Pointer diagnostics in
`internal/pluginhost/loader_windows.go` and its test, plus three copylocks
warnings in `internal/runtime/executor/kimi_executor.go` and
`internal/runtime/executor/oauth_scope_executor.go`. The raw diff of those four
files to the supplied base is empty (`pass8b-vet-unmodified-files.json`). No
baseline vet run was performed, so this is not a claimed reproduced baseline
result. No unrelated repairs were made. The fresh EXE build passed exactly once
at the new path and the binary was never executed.

An initial helper text-match alarm came from historical denial text inside the
preserved diff (git itself exited 0 with empty stderr), not a new permission
refusal. A premature read of this helper's own deliberately exclusive active
stdout hit a sharing lock; no alternative reader or bypass was used. Completed
receipts were read only after the known owned command handle finished. No
system-wide process survey was performed.

The original documentation prefix is preserved byte-for-byte. Final raw diff,
source hashes, protected-file/index/refs/runner checks, OFF state, build size/hash,
and final diff checks are recorded separately in
`pass8b-final-integrity.json` and `pass8b-validation-stage.json`; no
self-referential final diff hash is embedded in this appendix.

Remaining work is not waived: parent independent Standards/Spec review, the
six full-SPEC groups, baseline 0/4, Linux evidence, and the broader durability,
runtime/UI, clone/rebind, rollback and power-loss acceptance gates. In particular,
the vet gate is still red; passing Windows tests and compilation are not full
SPEC acceptance.