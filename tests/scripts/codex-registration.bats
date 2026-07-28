#!/usr/bin/env bats
#
# Tests for scripts/install/update-codex-marketplace-toml.py and the
# `make install-user-codex` target (DIR-037).
#
# Verifies that updating a Codex config.toml's
# [marketplaces.meta-cc-marketplace] source/source_type only touches those
# two keys and leaves all other TOML content byte-for-byte unchanged.
#
# Safety note: this test operates ONLY on synthetic fixture files under a
# temp dir (and a temp $HOME override for the Makefile target) -- it never
# reads or writes the real ~/.codex/config.toml.
#
# Run with: bats tests/scripts/codex-registration.bats

WORKTREE_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
SCRIPT="$WORKTREE_ROOT/scripts/install/update-codex-marketplace-toml.py"

write_fixture() {
    cat > "$1" <<'EOF'

[model_providers.bobdong]
name = "BobDong"
base_url = "https://bobdong.cn/v1"
env_key = "BOBDONG_API_KEY"
wire_api = "responses"

[projects."/home/yale/work/quay"]
trust_level = "trusted"

[tui.model_availability_nux]
"gpt-5.6-sol" = 4

[marketplaces.meta-cc-marketplace]
last_updated = "2026-07-27T02:06:56Z"
source_type = "local"
source = "/home/yale/work/meta-cc"

[plugins."meta-cc@meta-cc-marketplace"]
enabled = true

[model_providers.litellm]
name = "LiteLLM"
base_url = "https://litellm.lrfz.com"
env_key = "LITELLM_API_KEY"
wire_api = "responses"
EOF
}

setup() {
    FIXTURE="$BATS_TEST_TMPDIR/config.toml"
    write_fixture "$FIXTURE"
    cp "$FIXTURE" "$BATS_TEST_TMPDIR/config.toml.orig"
}

@test "update-codex-marketplace-toml.py updates source and leaves everything else byte-for-byte unchanged" {
    run python3 "$SCRIPT" "$FIXTURE" "/home/testuser/.local/share/meta-cc" local
    [ "$status" -eq 0 ]
    [[ "$output" == *"source: /home/yale/work/meta-cc -> /home/testuser/.local/share/meta-cc"* ]]

    # Exactly one line differs from the original.
    DIFF_OUTPUT="$(diff "$BATS_TEST_TMPDIR/config.toml.orig" "$FIXTURE" || true)"
    CHANGED_LINES="$(echo "$DIFF_OUTPUT" | grep -c '^[<>]' || true)"
    [ "$CHANGED_LINES" -eq 2 ]  # one removed ("<"), one added (">") -- a single line replaced

    grep -qF 'source = "/home/testuser/.local/share/meta-cc"' "$FIXTURE"
    grep -qF 'source_type = "local"' "$FIXTURE"
    grep -qF 'last_updated = "2026-07-27T02:06:56Z"' "$FIXTURE"

    for marker in '[model_providers.bobdong]' '[projects."/home/yale/work/quay"]' \
                  '[tui.model_availability_nux]' '[plugins."meta-cc@meta-cc-marketplace"]' \
                  '[model_providers.litellm]'; do
        grep -qF "$marker" "$FIXTURE"
    done

    # Result must still be valid TOML.
    run python3 -c "import tomllib; tomllib.load(open('$FIXTURE','rb'))"
    [ "$status" -eq 0 ]
}

@test "update-codex-marketplace-toml.py is idempotent when already pointing at the target" {
    python3 "$SCRIPT" "$FIXTURE" "/home/testuser/.local/share/meta-cc" local >/dev/null

    cp "$FIXTURE" "$BATS_TEST_TMPDIR/config.toml.after-first"
    run python3 "$SCRIPT" "$FIXTURE" "/home/testuser/.local/share/meta-cc" local
    [ "$status" -eq 0 ]
    [[ "$output" == *"already points at"* ]]
    diff "$BATS_TEST_TMPDIR/config.toml.after-first" "$FIXTURE"
}

@test "update-codex-marketplace-toml.py is a no-op when config.toml has no meta-cc-marketplace table" {
    NOMKT="$BATS_TEST_TMPDIR/config-no-marketplace.toml"
    cat > "$NOMKT" <<'EOF'
[model_providers.bobdong]
name = "BobDong"
EOF
    cp "$NOMKT" "$BATS_TEST_TMPDIR/config-no-marketplace.toml.orig"

    run python3 "$SCRIPT" "$NOMKT" "/home/testuser/.local/share/meta-cc" local
    [ "$status" -eq 0 ]

    diff "$BATS_TEST_TMPDIR/config-no-marketplace.toml.orig" "$NOMKT"
}

@test "update-codex-marketplace-toml.py is a no-op when config.toml does not exist" {
    MISSING="$BATS_TEST_TMPDIR/does-not-exist.toml"
    run python3 "$SCRIPT" "$MISSING" "/home/testuser/.local/share/meta-cc" local
    [ "$status" -eq 0 ]
    [ ! -f "$MISSING" ]
}

@test "update-codex-marketplace-toml.py refuses to touch a file that is not valid TOML" {
    BAD="$BATS_TEST_TMPDIR/broken.toml"
    printf '[marketplaces.meta-cc-marketplace\nsource = "unterminated\n' > "$BAD"
    cp "$BAD" "$BATS_TEST_TMPDIR/broken.toml.orig"

    run python3 "$SCRIPT" "$BAD" "/home/testuser/.local/share/meta-cc" local
    [ "$status" -eq 0 ]  # guarded no-op, not a hard failure

    diff "$BATS_TEST_TMPDIR/broken.toml.orig" "$BAD"
}

@test "make install-user-codex updates a fake HOME's config.toml without touching the real one" {
    FAKE_HOME="$BATS_TEST_TMPDIR/fake-home"
    mkdir -p "$FAKE_HOME/.codex" "$FAKE_HOME/.local/share/meta-cc"
    write_fixture "$FAKE_HOME/.codex/config.toml"
    cp "$FAKE_HOME/.codex/config.toml" "$FAKE_HOME/.codex/config.toml.orig"

    run make -C "$WORKTREE_ROOT" install-user-codex HOME="$FAKE_HOME"
    [ "$status" -eq 0 ]

    EXPECTED_SOURCE="$FAKE_HOME/.local/share/meta-cc"
    grep -qF "source = \"$EXPECTED_SOURCE\"" "$FAKE_HOME/.codex/config.toml"

    DIFF_OUTPUT="$(diff "$FAKE_HOME/.codex/config.toml.orig" "$FAKE_HOME/.codex/config.toml" || true)"
    CHANGED_LINES="$(echo "$DIFF_OUTPUT" | grep -c '^[<>]' || true)"
    [ "$CHANGED_LINES" -eq 2 ]
}

@test "make install-user-codex is a no-op when the fake HOME has no ~/.codex/config.toml" {
    FAKE_HOME="$BATS_TEST_TMPDIR/fake-home-no-codex"
    mkdir -p "$FAKE_HOME/.local/share/meta-cc"

    run make -C "$WORKTREE_ROOT" install-user-codex HOME="$FAKE_HOME"
    [ "$status" -eq 0 ]
    [ ! -e "$FAKE_HOME/.codex" ]
}
