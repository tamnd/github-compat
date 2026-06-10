// Package gate defines every conformance gate ID and its metadata.
package gate

// ID is a conformance gate identifier (e.g., "GH-A1").
type ID = string

// Category groups gates by client ecosystem.
type Category struct {
	Name  string
	Gates []Meta
}

// Meta describes a single gate.
type Meta struct {
	ID       ID
	Client   string
	Op       string
	Criterion string
}

// All gate IDs, grouped by category.
var (
	// Git transport
	G = []Meta{
		{"G1", "git", "HTTPS clone", "Clone succeeds; git fsck passes"},
		{"G2", "git", "SSH clone", "Clone succeeds; git fsck passes"},
		{"G3", "git", "HTTPS push", "Push succeeds; branch visible via REST"},
		{"G4", "git", "SSH push", "Push succeeds"},
		{"G5", "git", "Force push", "Succeeds"},
		{"G6", "git", "PR ref fetch", "refs/pull/1/head fetchable"},
		{"G7", "git", "Shallow clone", "--depth 1 succeeds; 1 commit"},
		{"G8", "gh", "pr checkout", "Branch checked out"},
		{"G9", "GCM", "HTTPS with GCM", "Clone succeeds without password prompt"},
		{"G10", "git", "Protocol v2", "version 2 in packet trace"},
		{"G11", "git-lfs", "LFS stub", "501 + JSON error body"},
		{"G12", "git", "Deploy key SSH", "Deploy key clone succeeds"},
	}

	// Authentication
	A = []Meta{
		{"A1", "all", "PAT classic", "GET /user with ghp_ returns login"},
		{"A2", "all", "PAT fine-grained", "GET /user with github_pat_ returns login"},
		{"A3", "gh", "X-OAuth-Scopes", "Header present (even empty)"},
		{"A4", "gh", "OAuth device flow", "gh auth login completes"},
		{"A5", "gh", "auth status", "Prints login and token type"},
		{"A6", "oct-js", "App JWT", "Installation token ghs_ returned"},
		{"A7", "oct-js", "Installation token", "API call with ghs_ succeeds"},
		{"A8", "all", "Deploy key", "SSH clone with deploy key"},
		{"A9", "GCM", "GCM credential", "Stored in OS keychain"},
		{"A10", "all", "Anonymous", "Public repo GET returns 200"},
		{"A11", "all", "OAuth web flow", "Code exchange returns gho_ token"},
		{"A12", "Renovate", "App auth", "ghs_ token via ghinstallation"},
	}

	// gh CLI
	GH = []Meta{
		{"GH-A1", "gh", "auth status", "Login printed; exit 0"},
		{"GH-A2", "gh", "repo view", "JSON with name"},
		{"GH-A3", "gh", "repo list", "Non-empty array"},
		{"GH-A4", "gh", "repo create", "Exit 0; repo created"},
		{"GH-A5", "gh", "issue list", "Array returned"},
		{"GH-A6", "gh", "issue create", "Exit 0; number printed"},
		{"GH-A7", "gh", "issue view", "JSON with title"},
		{"GH-A8", "gh", "pr list", "Array returned"},
		{"GH-A9", "gh", "pr create", "Exit 0; URL printed"},
		{"GH-A10", "gh", "pr view --json", "mergeable field present"},
		{"GH-A11", "gh", "pr checkout", "Branch checked out"},
		{"GH-A12", "gh", "pr diff", "Starts with diff --git"},
		{"GH-A13", "gh", "pr merge", "Exit 0; PR merged"},
		{"GH-A14", "gh", "pr review --approve", "Exit 0"},
		{"GH-A15", "gh", "pr checks", "Array (may be empty)"},
		{"GH-A16", "gh", "release create", "Exit 0"},
		{"GH-A17", "gh", "release upload", "Asset attached"},
		{"GH-A18", "gh", "api /user", "JSON with login"},
		{"GH-A19", "gh", "api --paginate", "All pages in one array"},
		{"GH-A20", "gh", "api graphql", "data.viewer.login set"},
		{"GH-A21", "gh", "search repos", "Results returned"},
		{"GH-A22", "gh", "label list", "Labels listed"},
	}

	// Octokit JavaScript
	OCT = []Meta{
		{"OCT-1", "@octokit/rest", "users.getAuthenticated()", "login set"},
		{"OCT-2", "@octokit/rest", "repos.get()", "id, default_branch set"},
		{"OCT-3", "@octokit/rest", "paginate issues", "Array; pagination works"},
		{"OCT-4", "@octokit/rest", "create + close issue", "number > 0; state closed"},
		{"OCT-5", "@octokit/graphql", "viewer { login }", "Login matches REST"},
		{"OCT-6", "@octokit/graphql", "repository(...)", "id and defaultBranchRef.name set"},
		{"OCT-7", "@octokit/rest", "rate-limit headers", "x-ratelimit-limit present"},
		{"OCT-8", "@octokit/rest", "ETag / 304", "Second identical GET returns 304"},
		{"OCT-9", "@octokit/webhooks", "webhook verify", "verifyAndReceive succeeds"},
		{"OCT-10", "@octokit/auth-app", "installation token", "ghs_ token; GET succeeds"},
		{"OCT-11", "@actions/github", "GITHUB_API_URL env", "Uses https://HOST/api/v3"},
	}

	// Python
	PY = []Meta{
		{"PY-1", "PyGitHub", "g.get_user()", "login set"},
		{"PY-2", "PyGitHub", "g.get_repo()", "name, default_branch set"},
		{"PY-3", "PyGitHub", "r.get_issues()", "Paginated list"},
		{"PY-4", "PyGitHub", "r.create_issue()", "number > 0"},
		{"PY-5", "PyGitHub", "issue.create_comment()", "id set"},
		{"PY-6", "PyGitHub", "issue.edit(state=closed)", "State becomes closed"},
		{"PY-7", "PyGitHub", "r.create_git_ref()", "No exception"},
		{"PY-8", "PyGitHub", "r.get_labels()", "No exception"},
		{"PY-9", "PyGitHub", "r.get_pulls()", "No exception"},
		{"PY-10", "PyGitHub", "g.get_rate_limit()", "core.limit > 0"},
		{"PY-11", "ghapi", "GhApi(gh_host=HOST)", "login set"},
		{"PY-12", "gidgethub", "gh.getitem(/user)", "login key present"},
		{"PY-13", "gidgethub", "webhook verify", "Event.from_http succeeds"},
		{"PY-14", "GitPython", "Repo.clone_from()", "Clone succeeds"},
		{"PY-15", "github3.py", "GitHubEnterprise().repository()", "full_name set"},
	}

	// Go clients
	GOG = []Meta{
		{"GOG-1", "go-github", "Users.Get(ctx, \"\")", "login set"},
		{"GOG-2", "go-github", "Repositories.Get()", "name, default_branch set"},
		{"GOG-3", "go-github", "Issues.ListByRepo paginated", "NextPage parsed"},
		{"GOG-4", "go-github", "Issues.Create()", "number > 0"},
		{"GOG-5", "go-github", "RateLimit.Get()", "Core.Limit > 0"},
		{"GOG-6", "go-github", "PullRequests.Get()", "Number, State set"},
		{"GOG-7", "go-github", "PullRequests.Merge()", "No error; state closed"},
		{"GOG-8", "go-github", "Checks.CreateCheckRun()", "ID > 0"},
		{"GOG-9", "httpcache", "ETag caching", "Second GET returns 304"},
		{"GOG-10", "githubv4", "Viewer query", "Login set"},
		{"GOG-11", "githubv4", "Repository query", "ID set"},
		{"GOG-12", "ghinstallation", "Installation token", "ghs_ token; GET succeeds"},
	}

	// Ruby / .NET / Java
	RB = []Meta{
		{"RB-1", "Octokit.rb", "client.user", "login set"},
		{"RB-2", "Octokit.rb", "client.repo()", "name, default_branch set"},
		{"RB-3", "Octokit.rb", "client.issues()", "Array returned"},
		{"RB-4", "Octokit.rb", "create + close issue", "number > 0; state closed"},
		{"RB-5", "Octokit.rb", "client.rate_limit", "limit > 0"},
		{"RB-6", "Octokit.rb", "client.pull_requests()", "Array returned"},
		{"RB-7", "Octokit.rb", "client.create_hook()", "Hook created"},
		{"RB-8", "Octokit.rb", "auto_paginate: true", "All pages fetched"},
	}
	CS = []Meta{
		{"CS-1", "Octokit.NET", "client.User.Current()", "Login set"},
		{"CS-2", "Octokit.NET", "client.Repository.Get()", "Name, DefaultBranch set"},
		{"CS-3", "Octokit.NET", "GetAllForRepository()", "List returned"},
		{"CS-4", "Octokit.NET", "client.Issue.Create()", "Number > 0"},
		{"CS-5", "Octokit.NET", "client.PullRequest.Get()", "Number, State set"},
		{"CS-6", "Octokit.NET", "GetRateLimits()", "Core.Limit > 0"},
		{"CS-7", "Octokit.NET", "Branch protection round-trip", "No error"},
		{"CS-8", "Octokit.NET", "Check.Run.Create()", "Id > 0"},
	}
	JAV = []Meta{
		{"JAV-1", "hub4j", "gh.checkApiUrlValidity()", "No exception"},
		{"JAV-2", "hub4j", "gh.getMyself()", "login set"},
		{"JAV-3", "hub4j", "gh.getRepository()", "name, defaultBranch set"},
		{"JAV-4", "hub4j", "Issues pagination", "All issues iterable"},
		{"JAV-5", "hub4j", "repo.createHook()", "Hook created"},
		{"JAV-6", "hub4j", "repo.createCommitStatus()", "No exception"},
		{"JAV-7", "hub4j", "repo.createCheckRun()", "id > 0"},
	}

	// Terraform / IaC
	TF = []Meta{
		{"TF-1", "Terraform", "data.github_repository", "default_branch set"},
		{"TF-2", "Terraform", "github_repository create+destroy", "No error"},
		{"TF-3", "Terraform", "github_branch_protection (GraphQL)", "Applied and destroyed"},
		{"TF-4", "Terraform", "github_repository_webhook", "Webhook created"},
		{"TF-5", "Terraform", "github_label", "Label created"},
		{"TF-6", "Terraform", "github_repository_deploy_key", "Deploy key created"},
		{"TF-7", "Terraform", "github_team + github_team_repository", "Team access granted"},
		{"TF-8", "Terraform", "App auth (app_auth block)", "ghs_ token; plan succeeds"},
	}
	ATL = []Meta{
		{"ATL-1", "Atlantis", "Server startup", "No startup error"},
		{"ATL-2", "Atlantis", "Webhook delivery", "X-Hub-Signature-256 validates"},
		{"ATL-3", "Atlantis", "atlantis plan comment", "Plan posted as PR comment"},
		{"ATL-4", "Atlantis", "atlantis apply with approval", "Commit status posted"},
	}

	// CI/CD
	JNK = []Meta{
		{"JNK-1", "Jenkins", "Test connection", "Credentials verified"},
		{"JNK-2", "Jenkins", "Multibranch pipeline", "Branches/PRs discovered"},
		{"JNK-3", "Jenkins", "Push webhook -> build", "Build starts"},
		{"JNK-4", "Jenkins", "Build result -> commit status", "Status visible on commit"},
	}
	RNV = []Meta{
		{"RNV-1", "Renovate", "Dry-run", "No auth errors"},
		{"RNV-2", "Renovate", "Dependency update", "Update branch created"},
		{"RNV-3", "Renovate", "Create PR", "PR visible; labels applied"},
		{"RNV-4", "Renovate", "Deduplicate PR", "No duplicate created"},
		{"RNV-5", "Renovate", "Auto-merge", "PR merged"},
	}
	ARGO = []Meta{
		{"ARGO-1", "ArgoCD", "HTTPS sync", "Sync succeeds"},
		{"ARGO-2", "ArgoCD", "SSH sync", "Sync succeeds"},
		{"ARGO-3", "ArgoCD", "Push webhook -> refresh", "App refresh triggered"},
		{"ARGO-4", "ArgoCD", "ApplicationSet GitHub generator", "Apps created per repo"},
	}
	FLUX = []Meta{
		{"FLUX-1", "Flux", "GitRepository HTTPS", "Reconciliation successful"},
		{"FLUX-2", "Flux", "GitRepository SSH", "Reconciliation successful"},
		{"FLUX-3", "Flux", "Push -> auto-reconcile", "Next reconciliation immediate"},
		{"FLUX-4", "Flux", "Commit status", "Status posted to commit"},
	}
	DRONE = []Meta{
		{"DRONE-1", "Drone", "OAuth login", "User authenticated"},
		{"DRONE-2", "Drone", "Repo activation", "Webhook created"},
		{"DRONE-3", "Drone", "Push -> build", "Build runs"},
		{"DRONE-4", "Drone", "Build result", "Commit status posted"},
	}

	// IDE / credential tools
	VSC = []Meta{
		{"VSC-1", "VS Code PR ext", "Sign in", "Token acquired; user shown"},
		{"VSC-2", "VS Code PR ext", "PR panel", "PR list displayed"},
		{"VSC-3", "VS Code PR ext", "PR detail view", "Title, description, files shown"},
		{"VSC-4", "VS Code PR ext", "PR diff view", "Diff visible"},
		{"VSC-5", "VS Code PR ext", "Add review comment", "Comment appears on PR"},
		{"VSC-6", "VS Code PR ext", "Approve PR", "Review state APPROVED"},
		{"VSC-7", "VS Code PR ext", "Merge PR", "PR merged"},
		{"VSC-8", "VS Code PR ext", "CI status shown", "Checks/statuses visible"},
	}
	JB = []Meta{
		{"JB-1", "JetBrains", "Add GHES server", "Token validates"},
		{"JB-2", "JetBrains", "Clone Repository", "Repos listed"},
		{"JB-3", "JetBrains", "Create PR", "PR created"},
		{"JB-4", "JetBrains", "PR list view", "PRs displayed"},
		{"JB-5", "JetBrains", "Build status", "CI status visible"},
	}
	GD = []Meta{
		{"GD-1", "GitHub Desktop", "Sign in", "Authenticated; repos listed"},
		{"GD-2", "GitHub Desktop", "Clone", "Clone succeeds"},
		{"GD-3", "GitHub Desktop", "Push", "Push succeeds"},
		{"GD-4", "GitHub Desktop", "Create PR", "PR created"},
		{"GD-5", "GitHub Desktop", "CI status", "Status visible"},
	}
	GCM = []Meta{
		{"GCM-1", "GCM", "HTTPS clone", "No prompt; clone succeeds"},
		{"GCM-2", "GCM", "Push", "Succeeds without prompting"},
		{"GCM-3", "GCM", "Re-acquire after erase", "Flow completes"},
		{"GCM-4", "GCM", "git credential fill", "Returns stored credential"},
	}
	GCO = []Meta{
		{"GCO-1", "git-credential-oauth", "HTTPS clone", "Token acquired; clone succeeds"},
		{"GCO-2", "git-credential-oauth", "Discovery doc", "/.well-known/oauth-authorization-server returns JSON"},
	}

	// Webhooks
	WH = []Meta{
		{"WH-1", "all", "Create webhook -> ping", "Ping received; HMAC validates"},
		{"WH-2", "all", "Open issue -> event", "issues.opened received"},
		{"WH-3", "all", "Push -> event", "push with correct SHA"},
		{"WH-4", "all", "Open PR -> event", "pull_request.opened received"},
		{"WH-5", "all", "Merge PR -> event", "pull_request.closed + merged:true"},
		{"WH-6", "all", "PR review -> event", "pull_request_review.submitted"},
		{"WH-7", "all", "Issue comment -> event", "issue_comment.created"},
		{"WH-8", "all", "Create branch -> event", "create, ref_type: branch"},
		{"WH-9", "all", "Create tag -> event", "create, ref_type: tag"},
		{"WH-10", "all", "Delete branch -> event", "delete, ref_type: branch"},
		{"WH-11", "Probot", "Full webhook + API", "Round-trip works"},
		{"WH-12", "@octokit/webhooks", "Invalid signature", "verifyAndReceive throws"},
	}

	// Releases / Gists
	REL = []Meta{
		{"REL-1", "gh", "gh release create", "Release created"},
		{"REL-2", "gh", "gh release upload", "Asset attached"},
		{"REL-3", "gh", "gh release list", "Releases listed"},
		{"REL-4", "GoReleaser", "Full release run", "Release with assets created"},
		{"REL-5", "all", "GET releases/latest", "Latest release returned"},
		{"REL-6", "semantic-release", "Full run", "Release created; PRs commented"},
	}
	GIST = []Meta{
		{"GIST-1", "gh", "gh gist list", "List returned (501 if deferred)"},
		{"GIST-2", "gh", "gh gist create", "Gist created (501 if deferred)"},
	}
)

// All returns every gate in every category.
func All() []Meta {
	var all []Meta
	for _, cat := range Categories() {
		all = append(all, cat.Gates...)
	}
	return all
}

// Categories returns all gate categories in display order.
func Categories() []Category {
	return []Category{
		{"git-transport", G},
		{"auth", A},
		{"gh-cli", GH},
		{"octokit-js", OCT},
		{"python", PY},
		{"go-clients", GOG},
		{"ruby", RB},
		{"dotnet", CS},
		{"java", JAV},
		{"terraform", TF},
		{"atlantis", ATL},
		{"jenkins", JNK},
		{"renovate", RNV},
		{"argocd", ARGO},
		{"flux", FLUX},
		{"drone", DRONE},
		{"vscode", VSC},
		{"jetbrains", JB},
		{"github-desktop", GD},
		{"gcm", GCM},
		{"git-credential-oauth", GCO},
		{"webhooks", WH},
		{"releases", REL},
		{"gists", GIST},
	}
}
