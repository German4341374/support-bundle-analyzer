# Redaction and privacy

`standard` replaces credential-like values, private keys, authorization values, tokens and connection strings. `strict` additionally pseudonymizes email addresses, IPv4 addresses, phone-like values and home-directory usernames.

Pseudonyms are stable only inside one export, which preserves correlation without exposing the original value. Binary files are excluded because byte-level rewriting can corrupt them and secret detection cannot reliably understand every binary format. The redaction manifest records source/output hashes, replacements by category and excluded files without recording matched values.

Always inspect `redaction-manifest.json` and the resulting text files before sharing. Redaction is risk reduction, not a proof that the export contains no sensitive data.
