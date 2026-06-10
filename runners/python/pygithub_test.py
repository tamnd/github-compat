#!/usr/bin/env python3
"""PyGitHub / Python ecosystem conformance runner. Gates PY-1 through PY-15."""

import os
import sys
import time
import asyncio
import traceback

HOST = os.environ.get("GITHOME_HOST") or sys.exit("need GITHOME_HOST")
TOKEN = os.environ.get("GITHOME_TOKEN") or sys.exit("need GITHOME_TOKEN")
OWNER = os.environ.get("GITHOME_OWNER", "test-owner")
REPO = os.environ.get("GITHOME_REPO", "test-repo")
ISSUE = int(os.environ.get("GITHOME_ISSUE_NUMBER", "1"))
PR = int(os.environ.get("GITHOME_PR_NUMBER", "1"))
WEBHOOK_SECRET = os.environ.get("GITHOME_WEBHOOK_SECRET", "test-secret")
INSECURE = os.environ.get("GITHOME_INSECURE") == "1"
SCHEME = "http" if INSECURE else "https"
REST_BASE = f"{SCHEME}://{HOST}/api/v3"

TOTAL = 15
results = []
n = [0]


def _pass(gate, desc):
    n[0] += 1
    results.append(f"ok {n[0]} - {gate}: {desc}")


def _fail(gate, desc, err=""):
    n[0] += 1
    results.append(f"not ok {n[0]} - {gate}: {desc}")
    if err:
        results.append(f"  # {str(err)[:200]}")


def _skip(gate, desc, reason):
    n[0] += 1
    results.append(f"ok {n[0]} - {gate}: {desc} # SKIP {reason}")


def run(gate, desc, fn, skip_reason=None):
    if skip_reason:
        _skip(gate, desc, skip_reason)
        return None
    try:
        result = fn()
        _pass(gate, desc)
        return result
    except Exception as e:
        _fail(gate, desc, e)
        return None


# --- PyGitHub ---
from github import Github, Auth

g = Github(auth=Auth.Token(TOKEN), base_url=REST_BASE)

# PY-1: g.get_user()
def py1():
    u = g.get_user()
    assert u.login, "no login"
run("PY-1", "g.get_user()", py1)

# PY-2: g.get_repo()
def py2():
    r = g.get_repo(f"{OWNER}/{REPO}")
    assert r.name, "no name"
    assert r.default_branch, "no default_branch"
run("PY-2", "g.get_repo()", py2)

# PY-3: r.get_issues() paginated
def py3():
    r = g.get_repo(f"{OWNER}/{REPO}")
    issues = list(r.get_issues(state="all"))
    assert isinstance(issues, list)
run("PY-3", "r.get_issues() paginated", py3)

# PY-4: r.create_issue()
def py4():
    r = g.get_repo(f"{OWNER}/{REPO}")
    iss = r.create_issue(title=f"compat-py4-{int(time.time())}", body="auto")
    assert iss.number > 0
    iss.edit(state="closed")
run("PY-4", "r.create_issue()", py4)

# PY-5: issue.create_comment()
def py5():
    r = g.get_repo(f"{OWNER}/{REPO}")
    iss = r.get_issue(ISSUE)
    c = iss.create_comment("compat test comment")
    assert c.id > 0
    c.delete()
run("PY-5", "issue.create_comment()", py5)

# PY-6: issue.edit(state="closed")
def py6():
    r = g.get_repo(f"{OWNER}/{REPO}")
    iss = r.create_issue(title=f"compat-py6-{int(time.time())}", body="auto")
    iss.edit(state="closed")
    iss2 = r.get_issue(iss.number)
    assert iss2.state == "closed"
run("PY-6", "issue.edit(state=closed)", py6)

# PY-7: r.create_git_ref()
def py7():
    import random, string
    r = g.get_repo(f"{OWNER}/{REPO}")
    branches = list(r.get_branches())
    assert branches, "no branches"
    sha = branches[0].commit.sha
    branch_name = "compat-py7-" + ''.join(random.choices(string.ascii_lowercase, k=6))
    ref = r.create_git_ref(f"refs/heads/{branch_name}", sha)
    assert ref.ref
    ref.delete()
run("PY-7", "r.create_git_ref()", py7)

# PY-8: r.get_labels()
def py8():
    r = g.get_repo(f"{OWNER}/{REPO}")
    labels = list(r.get_labels())
    assert isinstance(labels, list)
run("PY-8", "r.get_labels()", py8)

# PY-9: r.get_pulls()
def py9():
    r = g.get_repo(f"{OWNER}/{REPO}")
    prs = list(r.get_pulls(state="all"))
    assert isinstance(prs, list)
run("PY-9", "r.get_pulls()", py9)

# PY-10: g.get_rate_limit()
def py10():
    rl = g.get_rate_limit()
    assert rl.core.limit > 0
run("PY-10", "g.get_rate_limit()", py10)

# --- ghapi ---
try:
    from ghapi.all import GhApi as GhApiCls
    def py11():
        api = GhApiCls(token=TOKEN, gh_host=HOST)
        me = api.users.get_authenticated()
        assert me.login, "no login"
    run("PY-11", "ghapi GhApi(gh_host=HOST)", py11)
except ImportError as e:
    _skip("PY-11", "ghapi GhApi(gh_host=HOST)", f"ghapi not installed: {e}")

# --- gidgethub ---
try:
    import httpx
    import gidgethub.httpx as gh_httpx
    from gidgethub import sansio

    async def py12():
        async with httpx.AsyncClient() as client:
            gh = gh_httpx.GitHubAPI(client, "compat-test", oauth_token=TOKEN,
                                     base_url=REST_BASE + "/")
            data = await gh.getitem("/user")
            assert "login" in data, "no login key"

    run("PY-12", "gidgethub gh.getitem(/user)", lambda: asyncio.run(py12()))

    def py13():
        import hashlib, hmac, json
        secret = WEBHOOK_SECRET.encode()
        body = json.dumps({"action": "opened"}).encode()
        sig = "sha256=" + hmac.new(secret, body, hashlib.sha256).hexdigest()
        event = sansio.Event.from_http(
            {"X-GitHub-Event": "issues", "X-Hub-Signature-256": sig,
             "X-GitHub-Delivery": "abc-123"},
            body, secret=secret)
        assert event.event == "issues"
    run("PY-13", "gidgethub webhook verify", py13)
except ImportError as e:
    _skip("PY-12", "gidgethub gh.getitem(/user)", f"gidgethub not installed: {e}")
    _skip("PY-13", "gidgethub webhook verify", "gidgethub not installed")

# --- GitPython ---
try:
    import git as gitpython
    import tempfile, shutil
    def py14():
        d = tempfile.mkdtemp()
        try:
            url = f"{SCHEME}://x-token:{TOKEN}@{HOST}/{OWNER}/{REPO}.git"
            repo = gitpython.Repo.clone_from(url, d, depth=1)
            assert repo.head.commit
        finally:
            shutil.rmtree(d, ignore_errors=True)
    run("PY-14", "GitPython Repo.clone_from()", py14)
except ImportError as e:
    _skip("PY-14", "GitPython Repo.clone_from()", f"GitPython not installed: {e}")

# --- github3.py ---
try:
    import github3
    def py15():
        gh3 = github3.GitHubEnterprise(url=f"{SCHEME}://{HOST}", token=TOKEN)
        r = gh3.repository(OWNER, REPO)
        assert r is not None, "repository() returned None"
        assert r.full_name, "no full_name"
    run("PY-15", "github3.py GitHubEnterprise", py15)
except ImportError as e:
    _skip("PY-15", "github3.py GitHubEnterprise", f"github3.py not installed: {e}")

print("TAP version 14")
print(f"1..{TOTAL}")
for line in results:
    print(line)
