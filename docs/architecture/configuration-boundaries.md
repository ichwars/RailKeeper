# Configuration and maintenance boundaries

`configuration` owns personal preferences, digital-provider defaults, SMTP fallback/password
selection and maintenance sequencing through typed ports. `settingsdata` owns the existing
queries, UPSERT transactions, fixed database-file metrics and SQLite maintenance commands.
Application aliases preserve public records, methods and error identities; SMTP mailer
construction and transport stay in their existing adapter. No migration is introduced.

Partial personal updates merge into the same user's existing keys. Trimming and byte limits
remain unchanged. Validation and storage failures roll back the whole update; committed writes
remain committed if the final readback fails. Digital settings preserve provider defaults,
saved hosts/ports and postcommit reads without contacting hardware.

SMTP environment fallback applies only when no saved SMTP keys exist. Any saved SMTP record
retains the existing full-replacement semantics. Password bytes are never trimmed; blank input
retains the current password unless explicit clearing wins. Public responses contain only a
configured flag. `net/mail.ParseAddress` is allowed solely as a pure text parser in this module;
SMTP/network, database and filesystem dependencies remain forbidden in the independent code.

Maintenance retains checkpoint, VACUUM, optimize and final checkpoint, outside a transaction.
Native SQLite locking and the existing Admin/CSRF route remain in effect. File metrics use only
railkeeper.db and its WAL/SHM names under the configured directory. Missing files contribute
zero; reclaimed bytes never become negative. No production optimization is performed by tests.

Real SQLite fixtures cover private password state, preference isolation, validation/SQL/commit
rollback, cancelled operations, lost readback/retry and unchanged stored rows after optimization.
Automated mailer checks construct the existing adapter and send no external messages.
