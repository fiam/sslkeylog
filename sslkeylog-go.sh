#!/usr/bin/env bash
set -euo pipefail

# sslkeylog-go: wrap `go build|install` with an overlay that injects
#   import _ "github.com/fiam/sslkeylogfile/autopatch"
# into local `main` packages.
# No defaults, no flag parsing. You must provide {build|install} and all args.
# If no injection points are found, this exits with an error.

if (($# == 0)); then
  echo "usage: $(basename "$0") {build|install} [args...]" >&2
  exit 2
fi

subcmd="$1"; shift
case "$subcmd" in
  build|install) ;;
  *) echo "error: first arg must be 'build' or 'install'" >&2; exit 2 ;;
esac

# Identify candidate directories from args (dirs or files only).
declare -a cand_dirs=()
for tok in "$@"; do
  if [[ -d "$tok" ]]; then
    cand_dirs+=("${tok%/}")
  elif [[ -f "$tok" ]]; then
    cand_dirs+=("$(dirname "$tok")")
  fi
done

# De-duplicate dirs
if ((${#cand_dirs[@]} > 1)); then
  mapfile -t cand_dirs < <(printf '%s\n' "${cand_dirs[@]}" | awk '!seen[$0]++')
fi

# Helper: check if dir contains a `package main`
is_main_dir() {
  local d="$1"
  [[ -d "$d" ]] || return 1
  find "$d" -maxdepth 1 -type f -name '*.go' -print0 \
    | xargs -0 -r grep -qE '^[[:space:]]*package[[:space:]]+main([[:space:]]|$)'
}

# Filter only main dirs
declare -a main_dirs=()
for d in "${cand_dirs[@]}"; do
  if is_main_dir "$d"; then
    main_dirs+=("$d")
  fi
done

# Fail if no injection points found
if ((${#main_dirs[@]} == 0)); then
  echo "error: no local main packages found among provided args. Nothing to inject into." >&2
  exit 3
fi

echo "injecting sslkeylog autopatch into:"
for d in "${main_dirs[@]}"; do
  echo "  - $d"
done

# Create injected file and overlay
tmpdir="$(mktemp -d)"
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT

inject_src="$tmpdir/zz_sslkeylog_autopatch.go"
overlay="$tmpdir/overlay.json"

cat >"$inject_src" <<'EOF'
package main

import _ "github.com/fiam/sslkeylogfile/autopatch"
EOF

{
  echo '{'
  echo '  "Replace": {'
  for i in "${!main_dirs[@]}"; do
    [[ $i -gt 0 ]] && echo ','
    virt="${main_dirs[$i]%/}/zz_sslkeylog_autopatch.go"
    printf '    "%s": "%s"' "$virt" "$inject_src"
  done
  echo
  echo '  }'
  echo '}'
} >"$overlay"

echo "overlay file created at $overlay"

# Run go with overlay
exec go "$subcmd" -overlay="$overlay" "$@"
