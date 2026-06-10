# github-compat

Conformance test suite that proves any GitHub-compatible API host (Githome,
Gitea, Forgejo, GHES) works with every major client library, CLI tool, and
integration — without changing a line of client code.

[![CI](https://github.com/tamnd/github-compat/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/github-compat/actions/workflows/ci.yml)
[![Conformance](https://github.com/tamnd/github-compat/actions/workflows/conformance.yml/badge.svg)](https://github.com/tamnd/github-compat/actions/workflows/conformance.yml)

---

## What it tests

| Category | Gates | Clients |
|---|---|---|
| git transport | G1–G12 | git (HTTPS + SSH), protocol v2, LFS, deploy keys |
| Authentication | A1–A12 | PAT classic/fine-grained, OAuth device flow, GitHub Apps, GCM |
| gh CLI | GH-A1–GH-A22 | gh 2.72.0 — auth, repo, issue, pr, release, api, graphql |
| Octokit JS | OCT-1–OCT-11 | @octokit/rest 21.1.0, @octokit/graphql, @octokit/auth-app |
| Python | PY-1–PY-15 | PyGitHub 2.5.0, ghapi, gidgethub, GitPython, github3.py |
| Go clients | GOG-1–GOG-12 | go-github v66, githubv4, ghinstallation |
| Ruby | RB-1–RB-8 | octokit.rb 9.2.0 |
| .NET | CS-1–CS-8 | Octokit.NET 14.0.0 |
| Java | JAV-1–JAV-7 | hub4j/github-api 1.321 |
| Terraform | TF-1–TF-8 | terraform-provider-github 6.6.0 |
| Atlantis | ATL-1–ATL-4 | runatlantis/atlantis |
| Jenkins | JNK-1–JNK-4 | github-branch-source-plugin |
| Renovate | RNV-1–RNV-5 | renovatebot/renovate |
| ArgoCD | ARGO-1–ARGO-4 | argoproj/argo-cd |
| Flux | FLUX-1–FLUX-4 | fluxcd/flux2 |
| Drone / Woodpecker | DRONE-1–DRONE-4 | drone/drone, woodpecker-ci/woodpecker |
| VS Code | VSC-1–VSC-8 | vscode-pull-request-github |
| JetBrains | JB-1–JB-5 | IntelliJ GitHub plugin |
| GitHub Desktop | GD-1–GD-5 | desktop/desktop |
| GCM | GCM-1–GCM-4 | git-ecosystem/git-credential-manager |
| git-credential-oauth | GCO-1–GCO-2 | hickford/git-credential-oauth |
| Webhooks | WH-1–WH-12 | @octokit/webhooks, Probot, Jira, Linear |
| Releases | REL-1–REL-6 | gh, GoReleaser, semantic-release |
| Gists | GIST-1–GIST-2 | gh gist |

**Total: 191 gates**

---

## Quick start

```sh
# Build the conform binary
go build -o conform ./cmd/conform

# List all gates
./conform --list

# Run all runners against a live host
export GITHOME_HOST=githome.example.com
export GITHOME_TOKEN=ghp_yourtoken
export GITHOME_OWNER=my-org
export GITHOME_REPO=test-repo
./conform --report report.json --tap

# Run just the gh CLI suite
./conform --only gh-cli --tap

# Run just the Octokit JS suite
./conform --only oct-js --tap
```

---

## Environment variables

Every runner reads these from the environment:

| Variable | Required | Default | Description |
|---|---|---|---|
| `GITHOME_HOST` | yes | | Target host (no scheme, e.g. `githome.example.com`) |
| `GITHOME_TOKEN` | yes | | PAT with `repo`, `read:org`, `gist` scopes |
| `GITHOME_OWNER` | no | `test-owner` | Owner (user or org) of the test repo |
| `GITHOME_REPO` | no | `test-repo` | Repo name (must already exist with issues + a PR) |
| `GITHOME_ISSUE_NUMBER` | no | `1` | Pre-seeded open issue number |
| `GITHOME_PR_NUMBER` | no | `1` | Pre-seeded open PR number |
| `GITHOME_HEAD_SHA` | no | | HEAD commit SHA (auto-detected when absent) |
| `GITHOME_RELEASE_NAME` | no | `v0.0.1` | Pre-seeded release tag |
| `GITHOME_WEBHOOK_URL` | no | | Webhook delivery URL for WH-* gates |
| `GITHOME_WEBHOOK_SECRET` | no | `test-secret` | HMAC secret |
| `GITHOME_APP_ID` | no | | GitHub App ID (for A6, OCT-10, GOG-8, GOG-12) |
| `GITHOME_INSTALL_ID` | no | | Installation ID |
| `GITHOME_APP_PEM` | no | | RSA private key (inline PEM) |
| `GITHOME_SSH_HOST` | no | `$HOST:22` | SSH host override |
| `GITHOME_DEPLOY_KEY_PATH` | no | | Path to private key for G2/G4/G12 |
| `GITHOME_INSECURE` | no | `0` | Set to `1` to use HTTP instead of HTTPS |

---

## Seeding the fixture

The test repo must have:
- At least 1 commit on the default branch
- At least 1 open issue (set `GITHOME_ISSUE_NUMBER`)
- At least 1 open PR from a non-default branch (set `GITHOME_PR_NUMBER`)
- Labels: `bug`, `enhancement`, `documentation`
- At least 1 release (set `GITHOME_RELEASE_NAME`)

A seed script is in `runners/shell/seed.sh` (coming soon). You can also use
the gh CLI:

```sh
export GH_HOST=githome.example.com
export GH_ENTERPRISE_TOKEN=ghp_yourtoken

# Create test repo
gh repo create test-owner/test-repo --private --add-readme

# Create issue
gh issue create --repo test-owner/test-repo --title "First issue" --body "Seed issue"

# Create labels
gh label create bug --repo test-owner/test-repo --color d73a4a
gh label create enhancement --repo test-owner/test-repo --color a2eeef
gh label create documentation --repo test-owner/test-repo --color 0075ca

# Create a branch + PR
git clone https://githome.example.com/test-owner/test-repo.git
cd test-repo
git checkout -b feature/seed
echo "seed" >> seed.txt && git add seed.txt
git commit -m "seed"
git push origin feature/seed
gh pr create --repo test-owner/test-repo --title "Seed PR" --body "auto"

# Create a release
gh release create v0.0.1 --repo test-owner/test-repo --title v0.0.1 --notes "seed"
```

---

## CI/CD integration

### GitHub Actions

Set these repository secrets/variables:

| Name | Type | Description |
|---|---|---|
| `GITHOME_TOKEN` | Secret | PAT for the target host |
| `GITHOME_APP_PEM` | Secret | GitHub App private key |
| `GITHOME_HOST` | Variable | Target host |
| `GITHOME_OWNER` | Variable | Test org/user |
| `GITHOME_REPO` | Variable | Test repo name |
| `GITHOME_APP_ID` | Variable | GitHub App ID |
| `GITHOME_INSTALL_ID` | Variable | Installation ID |

Then trigger:

```sh
gh workflow run conformance.yml --field host=githome.example.com
```

Or the suite runs automatically on the daily schedule against the configured host.

### Differential testing

Compare Githome against api.github.com with a real org:

```sh
export GITHOME_HOST=githome.example.com
export GITHOME_TOKEN=ghp_githome_token
export GITHUB_REAL_TOKEN=ghp_dotcom_token
./conform --differential --report diff-report.json
```

Fields present in github.com responses but absent in Githome are hard failures.

---

## Runner reference

| Runner name | Language | Gates |
|---|---|---|
| `git` | bash | G1–G12 |
| `gh-cli` | bash | GH-A1–GH-A22 |
| `oct-js` | Node.js | OCT-1–OCT-11 |
| `python` | Python 3 | PY-1–PY-15 |
| `go-clients` | Go | GOG-1–GOG-12 |
| `ruby` | Ruby | RB-1–RB-8 |
| `dotnet` | C# | CS-1–CS-8 |
| `java` | Java | JAV-1–JAV-7 |
| `terraform` | bash + terraform | TF-1–TF-8 |

Run a single runner:

```sh
make run-gh        # gh CLI
make run-js        # Octokit JS
make run-python    # Python
make run-go        # Go clients
make run-ruby      # Ruby
make run-dotnet    # .NET
make run-java      # Java
make run-terraform # Terraform
```

---

## Gate naming

Gates follow the pattern `PREFIX-N`:

| Prefix | Category |
|---|---|
| `G` | git transport |
| `A` | authentication |
| `GH-A` | gh CLI |
| `OCT` | Octokit JS |
| `PY` | Python |
| `GOG` | Go clients |
| `RB` | octokit.rb |
| `CS` | Octokit.NET |
| `JAV` | hub4j Java |
| `TF` | Terraform |
| `ATL` | Atlantis |
| `JNK` | Jenkins |
| `RNV` | Renovate |
| `ARGO` | ArgoCD |
| `FLUX` | Flux |
| `DRONE` | Drone/Woodpecker |
| `VSC` | VS Code PR ext |
| `JB` | JetBrains |
| `GD` | GitHub Desktop |
| `GCM` | git-credential-manager |
| `GCO` | git-credential-oauth |
| `WH` | Webhooks |
| `REL` | Releases |
| `GIST` | Gists |

---

## TAP output

Every runner outputs [TAP version 14](https://testanything.org/tap-version-14-specification.html):

```
TAP version 14
1..22
ok 1 - GH-A1: auth status
ok 2 - GH-A2: repo view
not ok 3 - GH-A3: repo list
  # expected non-empty array, got []
ok 4 - GH-A4: repo create # SKIP not enough permissions
```

---

## Spec

The full compatibility spec lives at `~/notes/Spec/2001/compat/` — 16 documents
covering every endpoint, field, header, and behavior each client requires.
