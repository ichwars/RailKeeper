# Synthetic ECoS backup fixture

This pair was captured from the previous `syntheticAPIBackup` generator at
`0d24806f`. That generator used `golang.org/x/crypto/openpgp` v0.1.0-deprecated,
independently of the ProtonMail OpenPGP reader used by RailKeeper. The API test
reads the fixtures directly.

- `synthetic-backup.gpg` contains a gzip-compressed tar archive encrypted with
  AES-128 and protected by OpenPGP's modification detection code (MDC). The
  archive contains station serial `0x12345678` and two CV values for object 1002.
- `synthetic-keyring.gpg` is the matching keyring generated for this test using
  the identity `fixture@invalid.example`. The key material is disposable and
  public test data.
