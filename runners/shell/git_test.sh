#!/usr/bin/env bash
# git transport conformance runner — outputs TAP to stdout
# Gates: G1–G12

set -uo pipefail

HOST="${GITHOME_HOST:?need GITHOME_HOST}"
TOKEN="${GITHOME_TOKEN:?need GITHOME_TOKEN}"
OWNER="${GITHOME_OWNER:-test-owner}"
REPO="${GITHOME_REPO:-test-repo}"
SSH_HOST="${GITHOME_SSH_HOST:-$HOST}"

SCHEME="https"
[[ "${GITHOME_INSECURE:-}" == "1" ]] && SCHEME="http"

HTTPS_URL="$SCHEME://$HOST/$OWNER/$REPO.git"
SSH_URL="git@$SSH_HOST:$OWNER/$REPO.git"
HTTPS_CLONE_URL="${GITHOME_HTTPS_CLONE_URL:-$HTTPS_URL}"
DEPLOY_KEY="${GITHOME_DEPLOY_KEY_PATH:-}"

TOTAL=12
n=0
results=()

pass()    { n=$((n+1)); results+=("ok $n - $1: $2"); }
fail()    { n=$((n+1)); results+=("not ok $n - $1: $2"); [[ -n "${3:-}" ]] && results+=("  # $3"); }
skip_g()  { n=$((n+1)); results+=("ok $n - $1: $2 # SKIP $3"); }

WORK=$(mktemp -d /tmp/git-compat-XXXXXX)
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

# Helper: configure git credential for HTTPS
git_cred_url() {
  printf 'protocol=%s\nhost=%s\nusername=git\npassword=%s\n' "$SCHEME" "$HOST" "$TOKEN"
}
git config --global credential."$SCHEME://$HOST".username "x-token" 2>/dev/null || true

# G1: HTTPS clone
DIR="$WORK/g1"
out=$(GIT_ASKPASS=echo HTTPS_PROXY="" \
  git -c "credential.helper=" \
  -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
  clone "$HTTPS_CLONE_URL" "$DIR" 2>&1) && \
  git -C "$DIR" fsck --full 2>/dev/null && pass "G1" "HTTPS clone" || fail "G1" "HTTPS clone" "$out"

# G2: SSH clone
if [[ -n "$DEPLOY_KEY" ]]; then
  DIR2="$WORK/g2"
  out=$(GIT_SSH_COMMAND="ssh -i $DEPLOY_KEY -o StrictHostKeyChecking=no" \
    git clone "$SSH_URL" "$DIR2" 2>&1) && \
    git -C "$DIR2" fsck --full 2>/dev/null && pass "G2" "SSH clone" || fail "G2" "SSH clone" "$out"
else
  skip_g "G2" "SSH clone" "no GITHOME_DEPLOY_KEY_PATH set"
fi

# G3: HTTPS push
if [[ -d "$WORK/g1" ]]; then
  BRANCH="compat-push-$$"
  (cd "$WORK/g1" && \
    git checkout -b "$BRANCH" && \
    echo "compat $$" > compat_test.txt && \
    git add compat_test.txt && \
    git -c "user.email=compat@test" -c "user.name=Compat" commit -m "compat push test" && \
    git -c "credential.helper=" \
    -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
    push origin "$BRANCH" 2>&1) && pass "G3" "HTTPS push" || fail "G3" "HTTPS push" "push failed"
else
  skip_g "G3" "HTTPS push" "G1 clone failed"
fi

# G4: SSH push
if [[ -n "$DEPLOY_KEY" ]] && [[ -d "$WORK/g2" ]]; then
  BRANCH4="compat-ssh-push-$$"
  (cd "$WORK/g2" && \
    git checkout -b "$BRANCH4" && \
    echo "compat ssh" > compat_ssh.txt && \
    git add compat_ssh.txt && \
    git -c "user.email=compat@test" -c "user.name=Compat" commit -m "compat ssh push" && \
    GIT_SSH_COMMAND="ssh -i $DEPLOY_KEY -o StrictHostKeyChecking=no" \
    git push origin "$BRANCH4" 2>&1) && pass "G4" "SSH push" || fail "G4" "SSH push"
else
  skip_g "G4" "SSH push" "no deploy key or G2 clone failed"
fi

# G5: force push
if [[ -d "$WORK/g1" ]]; then
  FPBRANCH="compat-fp-$$"
  (cd "$WORK/g1" && \
    git checkout -b "$FPBRANCH" && \
    echo "first" > fp.txt && git add fp.txt && \
    git -c "user.email=compat@test" -c "user.name=Compat" commit -m "fp first" && \
    git -c "credential.helper=" \
    -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
    push origin "$FPBRANCH" && \
    echo "second" > fp.txt && git add fp.txt && \
    git -c "user.email=compat@test" -c "user.name=Compat" commit --amend --no-edit && \
    git -c "credential.helper=" \
    -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
    push origin "$FPBRANCH" --force 2>&1) && pass "G5" "force push" || fail "G5" "force push"
else
  skip_g "G5" "force push" "G1 clone failed"
fi

# G6: PR ref fetch
if [[ -d "$WORK/g1" ]]; then
  PR="${GITHOME_PR_NUMBER:-1}"
  out=$(cd "$WORK/g1" && \
    git -c "credential.helper=" \
    -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
    fetch origin "refs/pull/$PR/head" 2>&1) && pass "G6" "PR ref fetch" || fail "G6" "PR ref fetch" "$out"
else
  skip_g "G6" "PR ref fetch" "G1 clone failed"
fi

# G7: shallow clone
DIR7="$WORK/g7"
out=$(GIT_ASKPASS=echo \
  git -c "credential.helper=" \
  -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
  clone --depth 1 "$HTTPS_CLONE_URL" "$DIR7" 2>&1) && \
  COUNT=$(git -C "$DIR7" rev-list --count HEAD 2>/dev/null) && [[ "$COUNT" -eq 1 ]] && \
  pass "G7" "shallow clone" || fail "G7" "shallow clone" "$out"

# G8: skipped here (tested by gh CLI runner)
skip_g "G8" "pr checkout" "tested in gh-cli runner"

# G9: GCM (manual only)
skip_g "G9" "HTTPS with GCM" "requires interactive GCM setup"

# G10: protocol v2
DIR10="$WORK/g10"
out=$(GIT_TRACE_PACKET="$WORK/pkt.log" \
  git -c "credential.helper=" \
  -c "url.$SCHEME://x-token:$TOKEN@$HOST.insteadOf=$SCHEME://$HOST" \
  -c "protocol.version=2" \
  ls-remote "$HTTPS_CLONE_URL" 2>&1) && \
  grep -q 'version 2' "$WORK/pkt.log" 2>/dev/null && pass "G10" "Protocol v2" || fail "G10" "Protocol v2" "$out"

# G11: LFS stub
DIR11="$WORK/g11"
mkdir -p "$DIR11"
(cd "$DIR11" && git init -q && \
  git remote add origin "$HTTPS_CLONE_URL" && \
  # attempt a fake LFS batch request
  RESP=$(curl -sf -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/vnd.git-lfs+json" \
    -H "Accept: application/vnd.git-lfs+json" \
    -d '{"operation":"upload","transfers":["basic"],"objects":[{"oid":"abc","size":100}]}' \
    "$SCHEME://$HOST/$OWNER/$REPO.git/info/lfs/objects/batch" 2>&1 || true)
  HTTP_CODE=$(curl -so /dev/null -w '%{http_code}' -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/vnd.git-lfs+json" \
    -H "Accept: application/vnd.git-lfs+json" \
    -d '{"operation":"upload","transfers":["basic"],"objects":[{"oid":"abc","size":100}]}' \
    "$SCHEME://$HOST/$OWNER/$REPO.git/info/lfs/objects/batch" 2>/dev/null || echo "000")
  [[ "$HTTP_CODE" == "501" ]] || [[ "$HTTP_CODE" == "200" ]] ) && \
  pass "G11" "LFS stub" || fail "G11" "LFS stub" "expected 501 or 200"

# G12: deploy key SSH
if [[ -n "$DEPLOY_KEY" ]]; then
  DIR12="$WORK/g12"
  out=$(GIT_SSH_COMMAND="ssh -i $DEPLOY_KEY -o StrictHostKeyChecking=no" \
    git clone "$SSH_URL" "$DIR12" 2>&1) && pass "G12" "deploy key SSH" || fail "G12" "deploy key SSH" "$out"
else
  skip_g "G12" "deploy key SSH" "no GITHOME_DEPLOY_KEY_PATH"
fi

echo "TAP version 14"
echo "1..$TOTAL"
for line in "${results[@]}"; do
  echo "$line"
done
