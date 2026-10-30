# Resource catalog v1: opt-in inventory prerequisite

This is the first owner-side SPEC66/T01 slice, not a completed T01/SPEC66 or
SPEC79 gate implementation. It adds durable local source identity and read-only
inventory. It does not change inference eligibility, manual state, cooldowns,
refresh, balancing, scheduling, or the payload pipeline. There is no protection
ACK, pause/release operation, quota query, or gate write API.

## Enablement and support boundary

Existing deployments remain opted out. With the setting absent or disabled,
catalog initialization does not create state, issue markers, migrate sources, or
change routing. GET does not opt a deployment in.

For an owner-approved migration, configure the selected file explicitly:

```yaml
resource-catalog:
  enabled: true
  state-dir: .resource-catalog
# Keep the existing explicit oauth.auth-dir, matching the FileTokenStore directory.
```

Restart to apply enablement or a state-directory change. A relative state-dir is
resolved beside the selected configuration file; the default is
`.resource-catalog` there. The state directory and auth directory must not contain
one another. Persisted configuration must agree with the running settings.

Supported scope is standalone core with the actual auth manager using
`*auth.FileTokenStore`, on an operator-controlled local filesystem. The adapter
and resolved directory are checked rather than assumed from a global default.
Postgres, Git, object storage, custom store adapters, plugins, and Home are not
certified by this slice. Home retains its existing management unavailability and
no-local-fallback rules. No CLIProxyAPIHome protocol is changed or claimed ready.

File-object continuity is implemented on Windows (volume/file index and creation
time) and Linux (statx inode/device and birth time). Missing birth-time support,
unsupported platforms, symlinks in the auth tree, malformed sources, or I/O errors
do not produce a ready complete snapshot. The Windows implementation is tested
locally; a Linux cross-compile is not a Linux runtime/durability certification.
Network filesystems, distributed writers, volatile storage, and storage hardware
power-loss guarantees are outside the supported deployment contract. Operators
must not treat file-store type detection as certification of the storage substrate.

Enabling is an explicit source migration. Stop other catalog writers and source
editors during migration/recovery. Keep the entire state directory private and
persistent across restarts. Do not sync or version it as source code. Backups
contain the original credentials and need the same protection as the sources.

## HTTP read contract

Both routes use the existing management authentication and availability rules:

- `GET /v8/management/resource-capabilities`
- `GET /v8/management/resources`

No feature route is added to deprecated `/v0/management`. Missing authorization
uses the normal 401 behavior. Disabled management and Home preserve the normal
404 behavior. Responses use `Cache-Control: no-store`.

Capabilities return `version: 1`, `inventory`, `durable`, `gate`, `query`,
`policy_locked`, and `reason`. A current snapshot also supplies `scope` containing
the instance tuple, boot epoch, and snapshot revision. `gate` and `query` are
always false; `policy_locked` is always true because action policy is unsupported,
not because cooldowns or manual ownership have been inferred. Reasons are
`not_enabled`, `unsupported_mode`, `unsupported_storage`,
`unsupported_filesystem`, `not_ready`, or `available`. A storage/source failure
clears inventory/durable capability instead of pretending an old view is current.

Resources return a version-1 snapshot with `complete: true` only when all selected
sources and the durable state are current. Otherwise they return 503 with only:

```json
{"error":"resource_inventory_unavailable"}
```

DTOs are explicit allowlists. They contain issued IDs/versions, kind, a built-in
family enum (or `unsupported`), issued parent/member references, and separate
state axes:

- `manual.disabled`, `manual.excluded`, `manual.zero_weight`: source-local flags,
  including inherited group flags. These are not an overall eligibility result.
- `identity`: `valid` or `ambiguous`.
- `catalog_readiness`: `ready`, `not_ready`, or `unsupported`.
- `runtime_readiness`: always `unknown` in this slice.
- Per-row inventory capability, with query/gate false.

Source names, filenames, paths, endpoints, auth_index, credentials, raw Auth,
raw configuration, metadata, and private seals are not returned. An unknown or
unissued marker is represented as an ambiguous row without reflecting its
untrusted target string as an issued resource ID. A complete inventory can
contain ambiguous/unsupported rows; complete does not mean actionable or routable.

GET validates source and ledger freshness using reads only. It does not migrate,
refresh a token, contact an upstream, run a gate action, or write a journal. A
source change makes GET return unavailable until a successful local reconciliation.

## Source coverage and identity

The catalog reads the persisted normalized source tree before runtime flattening
or filtering. It includes all provider groups and configured key rows for Gemini,
Interactions, Vertex, Codex, Claude, xAI, Meta, and OpenAI compatibility, including
disabled compatibility providers, zero weight, exclusions, empty keys, and empty
groups. It recursively enumerates JSON auth envelopes just as FileTokenStore does.
Unknown auth families remain inventoried with explicit unsupported readiness;
plugin-expanded children are not claimed covered.

UUID-v4 identities are independently issued. IDs are not derived from, searched
by, or matched against names, paths, positions, auth_index, keys, or secret hashes.
Native/compatibility group and child envelopes receive `resource-id`; JSON auth
envelopes receive `resource_id`. Internal resource-group markers survive the
legacy flattened config boundary; reconstructing v8 groups uses those markers,
not endpoint/name matching. Filtered runtime rows are retained by typed saves
unless their explicit represented source identity was actually deleted.

Native and compatibility typed manual saves of opted-in or catalog-marked
source documents carry a transient, exact-source-document-bound projection for
each provider association and child. It is required even when provider IDs are
unique: child IDs can be duplicated within/across groups or absent. Only a source
entry positively represented in that loaded projection can be explicitly deleted;
filtered providers, keyless children and empty groups remain with their own source
groups. A duplicate provider ID never associates a filtered provider's credentials
with an enabled provider. Reordering, name/key/endpoint edits and represented
provider/child deletions remain supported.

The projection is not durable resource identity, never validates a duplicate ID,
and never infers identity by name, key, path or position. Group/child references
are checked in separate namespaces. A private in-memory document fence also covers
removed entries, which no longer carry child references: a stale save cannot erase
concurrent source additions by deleting an entire group or family. Unknown,
missing or stale typed provenance requires reloading the source, with no source
write. These values are excluded from JSON, stripped before source persistence,
and never enter the catalog, runtime Auth identity or diagnostic messages. An
ordinary unmarked opt-out document retains its existing YAML/runtime boundary.

The shared management persistence boundary preserves the existing array GET/PUT
and PATCH request shapes without exposing those private values. A loaded writer
retains an immutable receipt of its represented provider/child source references.
After the full document freshness check, JSON replacements can recover only a
unique source-marker address already present in that same writer/family receipt.
Compatibility GET does not return the provider marker: unique returned child
addresses can recover their one proven parent, never a parent inferred from its
name, endpoint, key or array position. Conflicting parents, duplicate addresses,
markers unknown to that receipt and stale documents are refused without changing source bytes.
Ordinary typed saves still require their explicit provenance; their existing
missing/unknown/cross-binding refusal rules are not relaxed.

This recovery is source context, not durable identity validation or issuance.
Credential changes still invalidate old credential/identity evidence; copied or
ambiguous catalog IDs do not become valid. Ambiguous duplicate child addresses
cannot reconstruct a unique source association through the existing JSON shape;
they require explicit source resolution or an existing provenance-bearing typed
writer, not a name/key/position guess. ID-less replacements/new entries stay
unmarked rather than guessing continuity when no unique address proves it. The
catalog retains sole responsibility for issuance, duplicate detection, physical
continuity, tombstones and observation fences. No endpoint handler, public JSON
shape, global serializer, writer registry or routing rule is added.

A successful owned typed save rebases only that writer's document fence and
represented row projections. Reconstruction records the exact output slots it
emitted, including new rows, reorders and deletions; it does not rediscover a
binding by ID, name, key or endpoint. Metadata is prepared before I/O and installed
only after persistence succeeds, under the caller's existing exclusive config
access (the manual handlers' mutex). No runtime-only or unrelated caller values
are replaced from a disk parse. A fresh `CloneForRuntime` carries the new basis;
older independent clones remain stale. Failed writes do not adopt a post-write
basis, and a subsequent external edit still requires reload, including when the
writer deletes every represented row. A row made runtime-filtered by an earlier
owned edit is not copied again as an untouched filtered source on the next save.
The receipt also retains its represented status after that successful transition:
removing it from the same writer, or a fresh runtime clone, is an explicit deletion
even though reparsing disk now filters it. The receipt is replaced only after a
successful write and never mutated through another clone. Rows never represented
by that writer, including external filtered rows and empty groups, remain preserved.

Native reconstruction retains shared group defaults on unrelated edits. Omitted
and null child fields retain their effective inheritance; an explicit empty list
or map remains an override. Where a typed edit deliberately clears an effective
value, reconstruction writes a non-null empty/zero child override. A cleared
nullable setting cannot express "use global" underneath a group override with the
existing parser: that group override is removed while other children retain their
materialized effective values. The runtime parser's inheritance semantics are
unchanged. This preserves both filtered-row semantics and legitimate manual
clears rather than restoring old settings after the save.

Clearing a native typed endpoint removes the old group endpoint even when the
empty field is omitted by YAML serialization. Reload uses the existing parser and
provider behavior: Gemini, Interactions, Vertex and Claude retain an empty
endpoint for provider fallback; Meta receives its parser default; Codex and xAI
remain filtered without a configured endpoint. Conflicting endpoints among
represented members of one source group are still rejected without a write. No
per-member endpoint grouping or compatibility-provider eligibility is invented.

Manual values use the runtime's decoding rules rather than raw scalar spelling.
For example, YAML `disabled: TRUE` is disabled; supported integer zero spellings
such as `0x0`, `0o0`, `0b0`, `+0`, and `-0` normalize to zero weight. Valid
non-positive weights also exclude weighted routing. Auth JSON retains its own
metadata semantics: disabled must be a JSON boolean, and weights use the shared
credential-weight parser (including numeric `0.0`/`0e0` and string `"0"`, but not
string `"0.0"`). Invalid values do not fabricate a ready inventory. A native
child's non-null exclusion list overrides inheritance, including `[]`; null keeps
the inherited list. These flags still do not establish overall routability.

Runtime Auth attributes carry source references only. They are unverified marker
references, not new scheduler identity, eligibility, or an enforcement signal.
The file store preserves a supplied source marker and rejects conflicting markers.

### Observation fences

A snapshot binds every target to:

- Persistent `instance_id` and `instance_incarnation`.
- Separate newly issued `boot_epoch` on every open.
- Monotonic `snapshot_revision` for source publication/rebinding.
- Resource ID/incarnation and monotonic identity, credential, configuration, and
  provider membership versions. Credential membership_version is zero.

An unchanged restart preserves persistent identities and versions. Proven
in-place rename/reorder with source markers preserves resource incarnation;
relevant config/snapshot fences advance. Effective inherited credentials are
sealed canonically so flattening headers is not mistaken for account replacement.
Native omitted/null child headers inherit; an explicit empty map overrides them.
Typed header normalization and effective empty/default endpoints are canonical,
including Meta's parser default. Compatibility children use provider-level
headers/endpoints, not ignored child overrides. A noncredential edit can advance
the configuration fence without replacing credential identity; real header,
endpoint or key changes still rotate credential/identity versions and incarnation.
JSON token/account sealing remains conservative and unchanged: no trusted refresh
continuity is inferred. Physical-file and source-freshness seals remain separate
from credential seals and do not gain any new continuity exception.

A key/account/token change conservatively advances resource incarnation plus
identity/credential versions. This slice has no trusted same-account refresh
attestation, so it never guesses that a changed OAuth token proves continuity.
Provider membership changes advance membership/config/identity fences. Removed
IDs are retained as tombstones; a reappearance is blocked, not resurrected by
matching a name or secret. New unmarked rows receive fresh issued identities.
Duplicate IDs block all affected observations and remain blocked after restart.

Private keyed seals detect change only; they never issue/match resource identity.
Physical-file seals distinguish a proven file rename from a byte-identical copied
replacement. Replacing a marked source with a new file object, including an
external editor's atomic-save operation, is conservatively ambiguous. Only the
catalog's own journaled marker publication authorizes its file replacements.
This intentionally trades convenience for conservative evidence fencing.

Full offline restoration that preserves both file objects and all metadata cannot
be distinguished from continuity. Unobserved delete/reinsert of a marker within
an unchanged envelope has the same inherent limitation. Do not reuse observations
across such operations. The module has an explicit version-checked Rebind primitive
that rotates instance incarnation, tested independently; no HTTP route, CLI
migration command, or automatic startup invokes it. Operational clone/restore
support still needs an owner-controlled procedure/tool. Do not blindly copy state
and claim another deployment is the same instance.

## Publication, concurrency, and recovery

The local catalog holds an OS-backed exclusive lock for its lifetime. Its mutex
serializes publication. Reconcile checks an expected full version tuple; stale,
cross-instance, and cross-boot calls conflict. Observation Check additionally
checks the resource target, tombstone, ambiguity, and current source/storage view.
A successful Check is NOT a routing admission or protection ACK.

Files in the state directory are private:

- `seal.key`: random change-detection key.
- `catalog.json`: issued identities, versions, private seals, and tombstones.
- `journal.json`: signed publication intent, source paths, marker slots, sealed
  preconditions and proposed state, not credential plaintext.
- `writer.lock`: lifetime writer coordination.

A fresh initialization persists an issued empty genesis ledger before attempting
source migration. It is not a ready snapshot: source validation must still
succeed. This lets a malformed first-start source be repaired and reopened
without treating an orphan key as continuity proof. An established ledger/key
mismatch or corruption remains unavailable; neither watcher reconciliation nor
reopen regenerates a missing member of an established ledger/key pair.

A marker migration first writes a recoverable intent. Original source backups are
written beside the sources as
`<source>.<instance_id>.<snapshot_revision>.catalog-v1-backup`.
Files are flushed and individually replaced; Unix replacements sync the directory,
and Windows replacements request write-through. This is **not a cross-file atomic
transaction**. The ledger is published only after source checks/replacements;
startup recovers a pending signed intent before exposing inventory.

Publication intent format 2 pins both content and physical identity for every
selected source, not only files being migrated. For a catalog-owned replacement,
the new file is prepared and flushed first; its physical identity is signed into
the journal before the rename. Recovery accepts only that authorized object.
Equal bytes are never a physical-continuity receipt. Replacing an untouched source
after collection, or copying an already published source while recovery is
pending, rejects publication/recovery rather than silently rebinding evidence.
Older pending format-1 intents lack this proof and are not silently upgraded.

Publication failure with a pending intent latches inventory unavailable until
restart recovery. Conflicting source edits leave the intent unresolved rather
than overwriting them. Missing/corrupt state or seal keys are not silently recreated
by the watcher. A transient pre-journal parse failure can be retried. Invalid,
unsupported, or read-only storage never yields fabricated readiness.

With a catalog hook installed before Start, the existing fsnotify watcher also
anchors the config's parent directory. It retains exact-path filtering and the
existing debounce/reload path, so owned marker replacements and external atomic
saves do not sever later manual edits/disablement. Directory-watch setup failure
is reported rather than hidden. Opt-out startup keeps its original watch setup.
This started-watcher sequence is verified on Windows; Linux runtime acceptance
remains outstanding.

The existing runtime fsnotify watcher watches the root auth directory. Recursive
sources are included and checked on every GET, but external nested-directory
edits are not promised automatic fsnotify publication by this slice. They remain
unavailable until an existing persisted-write hook, config reload, explicit local
reconciliation, or restart publishes them. No polling scheduler was added.

The lifetime lock coordinates catalog writers, not arbitrary old binaries/editors.
Sealed before/after checks reject observed external conflicts; they cannot turn
uncooperative multi-file writers into an atomic transaction. Keep such writers
stopped during migration and recovery.

## Migration and rollback limitations

1. Back up configuration, auth sources, and (if present) the complete state dir
   while stopped. Review the opt-in setting and directory ownership.
2. Start the opted-in supported core in the owner's approved deployment window.
   This task does not perform that deployment. Migration retains manual semantics
   and creates source backups without transferring any historical automatic owner.
3. Verify both inventory and row validity. Do not interpret available inventory
   as permission to enable future protection actions.
4. On failure, stop the process and retain sources, backups, journal, and key as
   one recovery set. Fix the I/O obstruction and restart; a source conflict needs
   owner reconciliation, not deletion of the journal or ledger to force success.

Disabling catalog on restart stops catalog operation without removing metadata or
rewriting manual/routing settings. Older strict parsers may reject the additive
setting/markers. There is no automatic downgrade exporter/removal command in this
slice. A rollback may restore pre-migration source backups only after accounting
for all subsequent manual/credential edits; do not overwrite newer credentials.
Retain the ledger rather than matching old identities after rollback. An approved
clone/restore requires fresh fences or explicit rebind, not a guessed identity.

No gate state exists yet, so this slice makes no safe-downgrade claim for a future
gate-enabled version. Such a version must separately prevent an older gate-ignorant
binary from routing and prove admission/retry/WebSocket behavior.

## Remaining acceptance

This is not full T01/SPEC66/SPEC79 completion. Unimplemented/unverified work includes:
independent gate ownership and durable ACKs; real admission/retry/continuation
coverage; provider quota/refresh/balance phases; Home/plugin/other-store support;
trusted same-account refresh attestations; operational clone/rebind and rollback
export tools; recursive external fsnotify publication; Linux runtime and actual
power-loss recovery; deployed-version/mode acceptance and coordinator review.

See [the verification record](resource-catalog-v1-verification.md) for exact local
commands, assertion red/green evidence, limitations, and final build outcomes.
