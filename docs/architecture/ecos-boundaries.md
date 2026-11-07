# ECoS model and transport ownership

`internal/ecosmodel` owns read-projection records, argument parsing, static attribute selection,
CV ordering, quoted comparison plans and copied telemetry records. It performs no network,
filesystem, HTTP or SQL operations. Shared CV byte/number limits belong to `internal/decoder`;
the application and protocol model use the same rules. Architecture tests enforce these imports.

The public application names remain aliases. JSON fields, stored profile text, CV pages,
source/conflict records, function snapshots and uncertain-write errors retain their contracts.
An absent CV index page remains distinct from an explicitly recorded zero page. Runtime speed,
direction and active function state remain excluded from the static locomotive projection.

| Application component | Responsibility |
| --- | --- |
| `ecos.go` | Service construction and the single shared connection/live-state owner |
| `ecos_connection.go` | Target normalization and connection/firmware read checks |
| `ecos_read.go` | Locomotive read/probe orchestration |
| `ecos_live.go`, `ecos_telemetry.go` | Existing cancellation, generations, locking, idle handling and sampling |
| `ecos_exchange.go` | Request/control/release sequencing and reply handling |
| `ecos_sync.go` | Preview or explicitly confirmed synchronization, including incomplete acknowledgments |
| `ecos_model_compat.go` | Compatibility names for existing callers and tests |

The transport implementation remains `internal/ecos`. There is still one `ECoSService`,
with the same mutexes, generation counters and connection lifetime. Splitting its methods
does not introduce a second monitor, client state, persistence path or background queue.

A comparison plan contains proposed values and command strings only. It cannot send them.
Synchronization remains a preview unless `confirm` is true and `dryRun` is false. Existing
workspace permissions, grants and explicit confirmations still gate actual operations.
Source loading, stored backups and image confirmation keep their existing specialized paths.

Verification includes the original simulated TCP/device and live-concurrency tests, direct
pure-rule tests for profile preservation, injection rejection, unknown/zero index pages and
non-aliasing telemetry copies, plus bounded quote round-trip fuzzing. These checks use local
simulated endpoints; they are not evidence of a new physical-device acceptance test.
This refactor introduces no migration or backup/transfer format change.
