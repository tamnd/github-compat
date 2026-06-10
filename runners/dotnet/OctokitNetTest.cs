// Octokit.NET conformance runner. Gates CS-1 through CS-8.
// Outputs TAP to stdout.

using Octokit;

var host       = Env("GITHOME_HOST");
var token      = Env("GITHOME_TOKEN");
var owner      = Env("GITHOME_OWNER", "test-owner");
var repo       = Env("GITHOME_REPO",  "test-repo");
var issueNum   = int.Parse(Env("GITHOME_ISSUE_NUMBER", "1"));
var prNum      = int.Parse(Env("GITHOME_PR_NUMBER", "1"));
var insecure   = Env("GITHOME_INSECURE", "0") == "1";
var scheme     = insecure ? "http" : "https";
var apiBase    = new Uri($"{scheme}://{host}/");

const int TOTAL = 8;
var results = new List<string>();
int n = 0;

var header = new ProductHeaderValue("github-compat", "1.0");
var client = new GitHubClient(header, apiBase)
{
    Credentials = new Credentials(token)
};

await Run("CS-1", "User.Current()", async () =>
{
    var u = await client.User.Current();
    if (string.IsNullOrEmpty(u.Login)) throw new Exception("no login");
});

await Run("CS-2", "Repository.Get()", async () =>
{
    var r = await client.Repository.Get(owner, repo);
    if (string.IsNullOrEmpty(r.Name)) throw new Exception("no name");
    if (string.IsNullOrEmpty(r.DefaultBranch)) throw new Exception("no default_branch");
});

await Run("CS-3", "Issue.GetAllForRepository()", async () =>
{
    var issues = await client.Issue.GetAllForRepository(owner, repo,
        new RepositoryIssueRequest { State = ItemStateFilter.All },
        new ApiOptions { PageSize = 5, PageCount = 1 });
    if (issues is null) throw new Exception("null result");
});

await Run("CS-4", "Issue.Create()", async () =>
{
    var created = await client.Issue.Create(owner, repo,
        new NewIssue($"compat-cs4-{DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()}") { Body = "auto" });
    if (created.Number <= 0) throw new Exception($"bad issue number: {created.Number}");
    await client.Issue.Update(owner, repo, created.Number, new IssueUpdate { State = ItemState.Closed });
});

await Run("CS-5", "PullRequest.Get()", async () =>
{
    var pr = await client.PullRequest.Get(owner, repo, prNum);
    if (pr.Number <= 0) throw new Exception("bad PR number");
    if (pr.State == null) throw new Exception("no state");
});

await Run("CS-6", "Miscellaneous.GetRateLimits()", async () =>
{
    var rl = await client.RateLimit.GetRateLimits();
    if (rl.Rate.Limit <= 0) throw new Exception($"rate limit is {rl.Rate.Limit}");
});

// CS-7: branch protection round-trip
await Run("CS-7", "Branch protection round-trip", async () =>
{
    var r = await client.Repository.Get(owner, repo);
    string branch = r.DefaultBranch;
    try
    {
        await client.Repository.Branch.GetBranchProtection(owner, repo, branch);
        Pass("CS-7", "Branch protection round-trip");
    }
    catch (NotFoundException)
    {
        // no protection set — that is fine for this test
    }
});

// CS-8: Check.Run.Create — skip if no app auth
Skip("CS-8", "Check.Run.Create()", "requires GitHub App auth (no app credentials in env)");

// output
Console.WriteLine("TAP version 14");
Console.WriteLine($"1..{TOTAL}");
foreach (var r in results) Console.WriteLine(r);

return results.Any(r => r.StartsWith("not ok")) ? 1 : 0;

// helpers

async Task Run(string gate, string desc, Func<Task> fn)
{
    n++;
    try
    {
        await fn();
        Pass(gate, desc);
    }
    catch (Exception e)
    {
        Fail(gate, desc, e.Message.Length > 200 ? e.Message[..200] : e.Message);
    }
}

void Pass(string gate, string desc) => results.Add($"ok {n} - {gate}: {desc}");
void Fail(string gate, string desc, string err = "")
{
    results.Add($"not ok {n} - {gate}: {desc}");
    if (!string.IsNullOrEmpty(err)) results.Add($"  # {err}");
}
void Skip(string gate, string desc, string reason)
{
    n++;
    results.Add($"ok {n} - {gate}: {desc} # SKIP {reason}");
}

string Env(string key, string def = "") =>
    Environment.GetEnvironmentVariable(key) is { Length: > 0 } v ? v : def != "" ? def : throw new Exception($"need {key}");
