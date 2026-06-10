#!/usr/bin/env bash
# gh CLI conformance runner — outputs TAP to stdout
# Gates: GH-A1 through GH-A22

set -uo pipefail

HOST="${GITHOME_HOST:?need GITHOME_HOST}"
TOKEN="${GITHOME_TOKEN:?need GITHOME_TOKEN}"
OWNER="${GITHOME_OWNER:-test-owner}"
REPO="${GITHOME_REPO:-test-repo}"
ISSUE="${GITHOME_ISSUE_NUMBER:-1}"
PR="${GITHOME_PR_NUMBER:-1}"

TOTAL=22
n=0
results=()

pass()    { n=$((n+1)); results+=("ok $n - $1: $2"); }
fail()    { n=$((n+1)); results+=("not ok $n - $1: $2"); [[ -n "${3:-}" ]] && results+=("  # $3"); }
skip_all(){ n=$((n+1)); results+=("ok $n - $1: $2 # SKIP $3"); }

# Configure gh to talk to the target host
export GH_HOST="$HOST"
export GH_ENTERPRISE_TOKEN="$TOKEN"
export GH_NO_UPDATE_NOTIFIER=1
GH="gh --hostname $HOST"

# GH-A1: auth status
out=$($GH auth status 2>&1) && pass "GH-A1" "auth status" || fail "GH-A1" "auth status" "$out"

# GH-A2: repo view
out=$($GH repo view "$OWNER/$REPO" --json name 2>&1) && \
  echo "$out" | grep -q '"name"' && pass "GH-A2" "repo view" || fail "GH-A2" "repo view" "$out"

# GH-A3: repo list
out=$($GH repo list "$OWNER" --json name --limit 5 2>&1) && \
  echo "$out" | grep -q '\[' && pass "GH-A3" "repo list" || fail "GH-A3" "repo list" "$out"

# GH-A4: repo create (idempotent: delete first if exists)
TMPNAME="compat-create-$$"
$GH repo delete "$OWNER/$TMPNAME" --yes 2>/dev/null || true
out=$($GH repo create "$OWNER/$TMPNAME" --private --description "compat test" 2>&1) && \
  pass "GH-A4" "repo create" || fail "GH-A4" "repo create" "$out"
$GH repo delete "$OWNER/$TMPNAME" --yes 2>/dev/null || true

# GH-A5: issue list
out=$($GH issue list --repo "$OWNER/$REPO" --json number --limit 5 2>&1) && \
  echo "$out" | grep -q '\[' && pass "GH-A5" "issue list" || fail "GH-A5" "issue list" "$out"

# GH-A6: issue create
out=$($GH issue create --repo "$OWNER/$REPO" --title "compat test $$" --body "automated" 2>&1) && \
  echo "$out" | grep -qE '(#[0-9]+|/issues/)' && pass "GH-A6" "issue create" || fail "GH-A6" "issue create" "$out"

# GH-A7: issue view
out=$($GH issue view "$ISSUE" --repo "$OWNER/$REPO" --json title 2>&1) && \
  echo "$out" | grep -q '"title"' && pass "GH-A7" "issue view" || fail "GH-A7" "issue view" "$out"

# GH-A8: pr list
out=$($GH pr list --repo "$OWNER/$REPO" --json number --limit 5 2>&1) && \
  echo "$out" | grep -q '\[' && pass "GH-A8" "pr list" || fail "GH-A8" "pr list" "$out"

# GH-A9: pr create — need a branch, skip if no non-main branch exists
if $GH api "repos/$OWNER/$REPO/git/refs" --hostname "$HOST" 2>/dev/null | grep -q '"feature/'; then
  out=$($GH pr create --repo "$OWNER/$REPO" --title "compat PR $$" --body "auto" --base main 2>&1) && \
    echo "$out" | grep -q 'https' && pass "GH-A9" "pr create" || fail "GH-A9" "pr create" "$out"
else
  skip_all "GH-A9" "pr create" "no feature branch exists"
fi

# GH-A10: pr view --json mergeable
out=$($GH pr view "$PR" --repo "$OWNER/$REPO" --json "number,mergeable,title" 2>&1) && \
  echo "$out" | grep -q '"number"' && pass "GH-A10" "pr view --json" || fail "GH-A10" "pr view --json" "$out"

# GH-A11: pr checkout (requires a local git repo, skip in pure CI)
if git rev-parse --git-dir >/dev/null 2>&1; then
  out=$($GH pr checkout "$PR" --repo "$OWNER/$REPO" 2>&1) && pass "GH-A11" "pr checkout" || fail "GH-A11" "pr checkout" "$out"
else
  skip_all "GH-A11" "pr checkout" "not in a git repo"
fi

# GH-A12: pr diff
out=$($GH pr diff "$PR" --repo "$OWNER/$REPO" 2>&1) && \
  echo "$out" | grep -q 'diff --git' && pass "GH-A12" "pr diff" || fail "GH-A12" "pr diff" "$out"

# GH-A13: pr merge — skip (would destroy fixture PR)
skip_all "GH-A13" "pr merge" "skipped to preserve fixture PR"

# GH-A14: pr review --approve — skip (needs separate PR)
skip_all "GH-A14" "pr review --approve" "skipped (needs separate PR)"

# GH-A15: pr checks
out=$($GH pr checks "$PR" --repo "$OWNER/$REPO" 2>&1); rc=$?
[[ $rc -eq 0 ]] || echo "$out" | grep -qi 'no checks' && \
  pass "GH-A15" "pr checks" || fail "GH-A15" "pr checks" "$out"

# GH-A16: release create
TAG="v0.0.compat-$$"
out=$($GH release create "$TAG" --repo "$OWNER/$REPO" --title "$TAG" --notes "compat test" 2>&1) && \
  pass "GH-A16" "release create" || fail "GH-A16" "release create" "$out"

# GH-A17: release upload
if $GH release view "$TAG" --repo "$OWNER/$REPO" >/dev/null 2>&1; then
  TMPFILE=$(mktemp /tmp/compat-asset-XXXXXX.bin)
  dd if=/dev/urandom bs=1024 count=4 of="$TMPFILE" 2>/dev/null
  out=$($GH release upload "$TAG" "$TMPFILE" --repo "$OWNER/$REPO" 2>&1) && \
    pass "GH-A17" "release upload" || fail "GH-A17" "release upload" "$out"
  rm -f "$TMPFILE"
else
  skip_all "GH-A17" "release upload" "release not created"
fi
# cleanup
$GH release delete "$TAG" --repo "$OWNER/$REPO" --yes 2>/dev/null || true

# GH-A18: api /user
out=$($GH api user --hostname "$HOST" 2>&1) && \
  echo "$out" | grep -q '"login"' && pass "GH-A18" "api /user" || fail "GH-A18" "api /user" "$out"

# GH-A19: api --paginate
out=$($GH api "repos/$OWNER/$REPO/issues?state=all&per_page=1" --paginate --hostname "$HOST" 2>&1) && \
  echo "$out" | grep -q '"number"' && pass "GH-A19" "api --paginate" || fail "GH-A19" "api --paginate" "$out"

# GH-A20: api graphql
QUERY='{ "query": "{ viewer { login } }" }'
out=$($GH api graphql --hostname "$HOST" -f query='{ viewer { login } }' 2>&1) && \
  echo "$out" | grep -q '"login"' && pass "GH-A20" "api graphql" || fail "GH-A20" "api graphql" "$out"

# GH-A21: search repos
out=$($GH search repos "" --hostname "$HOST" --json fullName --limit 3 2>&1); rc=$?
[[ $rc -eq 0 ]] && pass "GH-A21" "search repos" || fail "GH-A21" "search repos" "$out"

# GH-A22: label list
out=$($GH label list --repo "$OWNER/$REPO" 2>&1) && \
  pass "GH-A22" "label list" || fail "GH-A22" "label list" "$out"

echo "TAP version 14"
echo "1..$TOTAL"
for line in "${results[@]}"; do
  echo "$line"
done
