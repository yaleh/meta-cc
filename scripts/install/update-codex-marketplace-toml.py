#!/usr/bin/env python3
"""update-codex-marketplace-toml.py -- DIR-037

Update the `source` / `source_type` keys of the
`[marketplaces.meta-cc-marketplace]` table in a Codex `config.toml` file,
leaving every other line in the file byte-for-byte unchanged.

Why not a full TOML parse -> modify -> serialize round trip? A
general-purpose TOML writer would risk reformatting comments, blank lines,
key ordering, quoting style, or unrelated tables ([model_providers.*],
[projects.*], [plugins.*], ...) elsewhere in a developer's hand-edited
~/.codex/config.toml. Instead this performs a targeted, line-level edit:

  1. Parse the file with `tomllib` (read-only) purely to confirm it is
     valid TOML and to confirm the `[marketplaces.meta-cc-marketplace]`
     table actually exists. If the file is missing, unreadable, not valid
     TOML, or has no such table, this is a documented no-op (exit 0).
  2. Locate that table's header line and its end (the next `[...]` header
     or EOF) by simple line scanning.
  3. Within only that line range, find the `source_type = "..."` and
     `source = "..."` lines and rewrite just their values in place. Every
     other line -- inside or outside the table -- is copied through
     unmodified.
  4. Re-parse the edited content with `tomllib` to confirm the result is
     still valid TOML and has the expected values before committing it.
  5. Write the result to a temp file in the same directory and atomically
     rename it over the original (same safety pattern as `make stage`'s
     binary copy fix in this same task).

Usage:
  update-codex-marketplace-toml.py <config_toml_path> <new_source> [new_source_type]

new_source_type defaults to "local" (Codex's convention for a filesystem
path source, matching what `make install-user` already generates as the
directory-source marketplace entry).

Exit status is always 0 for "file not found" / "table not found" (these are
guarded no-ops per the DIR-037 contract: additive safety, not a required
step), and non-zero only for a genuine failure (invalid TOML input, or the
edit somehow produced invalid TOML output).
"""
import os
import re
import sys
import tempfile

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover - guarded, see main()
    tomllib = None

TABLE_HEADER_RE = re.compile(
    r'^\s*\[marketplaces\.(?:"meta-cc-marketplace"|meta-cc-marketplace)\]\s*$'
)
ANY_TABLE_HEADER_RE = re.compile(r"^\s*\[")


def find_table_bounds(lines):
    """Return (start, end) line-index range [start, end) of the
    [marketplaces.meta-cc-marketplace] table, or (None, None) if absent."""
    start = None
    for i, line in enumerate(lines):
        if TABLE_HEADER_RE.match(line):
            start = i
            break
    if start is None:
        return None, None
    end = len(lines)
    for i in range(start + 1, len(lines)):
        if ANY_TABLE_HEADER_RE.match(lines[i]):
            end = i
            break
    return start, end


def replace_key_in_table(lines, start, end, key, new_value):
    """Rewrite `key = "..."` within lines[start:end] to new_value, or insert
    it just before the end of the table block if not already present.
    Returns (old_value_or_None, changed_bool).

    Note: the line ending is stripped before matching and re-appended
    afterwards exactly as found (rather than folding it into the regex's
    trailing `\\s*`), because `\\s` also matches `\\n` -- matching it inline
    previously caused the newline to be captured AND re-appended, doubling
    it and corrupting the file with a spurious blank line."""
    key_re = re.compile(r'^(\s*' + re.escape(key) + r'\s*=\s*)(".*?")([ \t]*(?:#.*)?)$')
    for i in range(start, end):
        raw = lines[i]
        has_newline = raw.endswith("\n")
        body = raw[:-1] if has_newline else raw
        m = key_re.match(body)
        if m:
            old_value = m.group(2).strip('"')
            if old_value == new_value:
                return old_value, False
            new_body = f'{m.group(1)}"{new_value}"{m.group(3)}'
            lines[i] = new_body + ("\n" if has_newline else "")
            return old_value, True

    insert_at = end
    while insert_at > start + 1 and lines[insert_at - 1].strip() == "":
        insert_at -= 1
    lines.insert(insert_at, f'{key} = "{new_value}"\n')
    return None, True


def main(argv):
    if len(argv) < 3:
        print(
            "Usage: update-codex-marketplace-toml.py <config.toml> <new_source> [new_source_type]",
            file=sys.stderr,
        )
        return 2

    config_path = argv[1]
    new_source = argv[2]
    new_source_type = argv[3] if len(argv) > 3 else "local"

    if not os.path.isfile(config_path):
        print(f"  (no {config_path} found, skipping Codex registration)")
        return 0

    if tomllib is None:
        print(
            f"  (python3 tomllib module unavailable (needs Python 3.11+); "
            f"skipping Codex registration for {config_path})",
            file=sys.stderr,
        )
        return 0

    with open(config_path, "rb") as f:
        try:
            parsed = tomllib.load(f)
        except tomllib.TOMLDecodeError as exc:
            print(
                f"  (WARNING: {config_path} is not valid TOML ({exc}); "
                f"skipping Codex registration)",
                file=sys.stderr,
            )
            return 0

    if "meta-cc-marketplace" not in parsed.get("marketplaces", {}):
        print(
            f"  ([marketplaces.meta-cc-marketplace] not found in {config_path}, "
            f"skipping Codex registration)"
        )
        return 0

    with open(config_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    start, end = find_table_bounds(lines)
    if start is None:
        # Should not happen given the tomllib check above, but stay a no-op
        # rather than guessing at file structure.
        print(
            f"  ([marketplaces.meta-cc-marketplace] table header not found in "
            f"{config_path}, skipping Codex registration)"
        )
        return 0

    old_source, source_changed = replace_key_in_table(lines, start, end, "source", new_source)

    # source_type insertion (if it were missing) could shift line numbers,
    # so bounds are recomputed before the second edit for safety.
    start2, end2 = find_table_bounds(lines)
    old_source_type, source_type_changed = replace_key_in_table(
        lines, start2, end2, "source_type", new_source_type
    )

    if not source_changed and not source_type_changed:
        print(
            f"  {config_path} [marketplaces.meta-cc-marketplace] already "
            f"points at {new_source} (source_type: {new_source_type}); no change needed"
        )
        return 0

    new_content = "".join(lines)
    try:
        tomllib.loads(new_content)
    except tomllib.TOMLDecodeError as exc:
        print(
            f"  ERROR: edit would produce invalid TOML ({exc}); "
            f"leaving {config_path} untouched",
            file=sys.stderr,
        )
        return 1

    dir_name = os.path.dirname(os.path.abspath(config_path)) or "."
    fd, tmp_path = tempfile.mkstemp(prefix=".config.toml.", dir=dir_name)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(new_content)
        os.replace(tmp_path, config_path)
    except Exception:
        os.unlink(tmp_path)
        raise

    if source_changed:
        if old_source is None:
            print(
                f"✓ Added {config_path} [marketplaces.meta-cc-marketplace] source: {new_source}"
            )
        else:
            print(
                f"✓ Updated {config_path} [marketplaces.meta-cc-marketplace] "
                f"source: {old_source} -> {new_source}"
            )
    if source_type_changed:
        if old_source_type is None:
            print(
                f"✓ Added {config_path} [marketplaces.meta-cc-marketplace] "
                f"source_type: {new_source_type}"
            )
        else:
            print(
                f"✓ Updated {config_path} [marketplaces.meta-cc-marketplace] "
                f"source_type: {old_source_type} -> {new_source_type}"
            )

    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
